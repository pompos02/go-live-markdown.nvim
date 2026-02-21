// Package httpserver handles all message traffic between Neovim and the browser.
package httpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-live-markdown/internal/contracts"

	"github.com/gorilla/websocket"
)

type renderPayload struct {
	html     string
	filename string
}

// PreviewServer coordinates HTTP serving and WebSocket updates.
type PreviewServer struct {
	addr  string
	shell string

	started bool
	server  *http.Server

	// OnGoToLine is invoked when the browser requests a jump to a source line.
	OnGoToLine     func(contracts.GoToLineMessage)
	browserInbound chan []byte

	updates    chan renderPayload
	cursors    chan contracts.CursorMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	stopLoop   chan struct{}

	upgrader websocket.Upgrader
}

// NewPreviewServer creates an HTTP/WebSocket preview server bound to addr.
func NewPreviewServer(addr string, shell string) *PreviewServer {
	return &PreviewServer{
		addr:  addr,
		shell: shell,

		browserInbound: make(chan []byte, 64),
		updates:        make(chan renderPayload, 8),
		cursors:        make(chan contracts.CursorMessage, 32),
		register:       make(chan *websocket.Conn),
		unregister:     make(chan *websocket.Conn),
		stopLoop:       make(chan struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// URL returns the browser URL for the preview server.
func (m *PreviewServer) URL() string {
	return "http://" + m.addr
}

// StartOrUpdate starts the preview server on first call and publishes new HTML.
func (m *PreviewServer) StartOrUpdate(fragment string, path string) error {
	if !m.started {
		mux := http.NewServeMux()
		mux.HandleFunc("/", m.handleIndex)
		mux.HandleFunc("/ws", m.handleWS)
		mux.HandleFunc("/@mdfs/", m.handleAsset)

		m.server = &http.Server{Addr: m.addr, Handler: mux}
		m.started = true

		go m.runLoop()
		go func() {
			_ = m.server.ListenAndServe()
		}()
	}

	filename := filepath.Base(path)
	m.updates <- renderPayload{html: fragment, filename: filename}
	return nil
}

// UpdateCursor publishes a cursor update to connected browsers.
func (m *PreviewServer) UpdateCursor(msg contracts.CursorMessage) error {
	if !m.started {
		return nil
	}

	msg.Type = contracts.MessageTypeCursor
	m.cursors <- msg
	return nil
}

// Stop gracefully shuts down the HTTP server and run loop.
func (m *PreviewServer) Stop() error {
	if !m.started || m.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := m.server.Shutdown(ctx)

	close(m.stopLoop)

	m.started = false
	m.server = nil
	return err
}

// handleIndex serves the initial HTML shell.
func (m *PreviewServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(m.shell))
}

// handleWS upgrades the connection and forwards browser messages to the loop.
func (m *PreviewServer) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	m.register <- conn
	defer func() {
		m.unregister <- conn
	}()

	// Block here until the connection closes / errors outs
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		m.browserInbound <- msg
	}
}

// SetGoToLineHandler registers the callback for browser go-to-line requests.
func (m *PreviewServer) SetGoToLineHandler(fn func(contracts.GoToLineMessage)) {
	m.OnGoToLine = fn
}

// handleAsset serves local markdown assets via encoded absolute paths.
func (m *PreviewServer) handleAsset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/@mdfs/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	decoded, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	assetPath := filepath.Clean(string(decoded))
	if assetPath == "." || !filepath.IsAbs(assetPath) {
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(assetPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, assetPath)
}

// runLoop serializes state updates and websocket writes on a single goroutine.
func (m *PreviewServer) runLoop() {
	var conn *websocket.Conn

	lastRender := contracts.RenderMessage{Type: contracts.MessageTypeRender}
	lastCursor := contracts.CursorMessage{Type: contracts.MessageTypeCursor}
	haveCursor := false

	for {
		select {
		case update := <-m.updates:
			lastRender.Rev++
			lastRender.HTML = update.html
			lastRender.Filename = update.filename

			if conn == nil {
				continue
			}

			if !writeJSON(conn, lastRender) {
				conn = nil
				continue
			}

			if haveCursor {
				lastCursor.Rev = lastRender.Rev
				if !writeJSON(conn, lastCursor) {
					conn = nil
				}
			}

		case cursor := <-m.cursors:
			lastCursor = cursor
			haveCursor = true

			if conn == nil || lastRender.Rev == 0 {
				continue
			}

			lastCursor.Rev = lastRender.Rev
			if !writeJSON(conn, lastCursor) {
				conn = nil
			}

		case c := <-m.register:
			if conn != nil {
				_ = conn.Close()
			}
			conn = c

			if !writeJSON(conn, lastRender) {
				conn = nil
				continue
			}

			if haveCursor && lastRender.Rev > 0 {
				lastCursor.Rev = lastRender.Rev
				if !writeJSON(conn, lastCursor) {
					conn = nil
				}
			}

		case c := <-m.unregister:
			if conn == c {
				_ = conn.Close()
				conn = nil
			}

		case raw := <-m.browserInbound:
			var envelope contracts.IncomingMessage
			if err := json.Unmarshal(raw, &envelope); err != nil {
				continue
			}
			switch envelope.Type {
			case contracts.MessageTypeGoToLine:
				var msg contracts.GoToLineMessage
				if err := json.Unmarshal(raw, &msg); err != nil {
					continue
				}
				if m.OnGoToLine != nil {
					m.OnGoToLine(msg)
				}
			}

		case <-m.stopLoop:
			if conn != nil {
				_ = conn.Close()
				conn = nil
			}
			return
		}
	}
}

// writeJSON writes a JSON message and reports whether the connection is usable.
func writeJSON(conn *websocket.Conn, v any) bool {
	if err := conn.WriteJSON(v); err != nil {
		_ = conn.Close()
		return false
	}
	return true
}
