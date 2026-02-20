// Package httptransport handles all message traffic between Neovim and the browser.
package httptransport

import (
	"context"
	"net/http"
	"time"

	"go-live-markdown/internal/contracts"

	"github.com/gorilla/websocket"
)

type renderPayload struct {
	html string
}

// Manager coordinates HTTP serving and WebSocket updates.
type Manager struct {
	addr  string
	shell string

	started bool
	server  *http.Server

	updates    chan renderPayload
	cursors    chan contracts.CursorMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	stopLoop   chan struct{}

	upgrader websocket.Upgrader
}

func NewManager(addr string, shell string) *Manager {
	return &Manager{
		addr:       addr,
		shell:      shell,
		updates:    make(chan renderPayload, 8),
		cursors:    make(chan contracts.CursorMessage, 32),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		stopLoop:   make(chan struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (m *Manager) URL() string {
	return "http://" + m.addr
}

func (m *Manager) StartOrUpdate(fragment string) error {
	if !m.started {
		mux := http.NewServeMux()
		mux.HandleFunc("/", m.handleIndex)
		mux.HandleFunc("/ws", m.handleWS)

		m.server = &http.Server{Addr: m.addr, Handler: mux}
		m.started = true

		go m.runLoop()
		go func() {
			_ = m.server.ListenAndServe()
		}()
	}

	m.updates <- renderPayload{html: fragment}
	return nil
}

func (m *Manager) UpdateCursor(msg contracts.CursorMessage) error {
	if !m.started {
		return nil
	}

	msg.Type = contracts.MessageTypeCursor
	m.cursors <- msg
	return nil
}

func (m *Manager) Stop() error {
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

func (m *Manager) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(m.shell))
}

func (m *Manager) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	m.register <- conn
	defer func() {
		m.unregister <- conn
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (m *Manager) runLoop() {
	var conn *websocket.Conn

	lastRender := contracts.RenderMessage{Type: contracts.MessageTypeRender}
	lastCursor := contracts.CursorMessage{Type: contracts.MessageTypeCursor}
	haveCursor := false

	for {
		select {
		case update := <-m.updates:
			lastRender.Rev++
			lastRender.HTML = update.html

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

		case <-m.stopLoop:
			if conn != nil {
				_ = conn.Close()
				conn = nil
			}
			return
		}
	}
}

func writeJSON(conn *websocket.Conn, v any) bool {
	if err := conn.WriteJSON(v); err != nil {
		_ = conn.Close()
		return false
	}
	return true
}
