package httptransport

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type renderMessage struct {
	Type string `json:"type"`
	HTML string `json:"html"`
}

type Manager struct {
	addr  string
	shell string

	started bool
	server  *http.Server

	updates    chan string
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	stopLoop   chan chan struct{}

	upgrader websocket.Upgrader
}

func NewManager(addr string, shell string) *Manager {
	return &Manager{
		addr:       addr,
		shell:      shell,
		updates:    make(chan string, 8),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		stopLoop:   make(chan chan struct{}),
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

		m.server = &http.Server{
			Addr:    m.addr,
			Handler: mux,
		}

		m.started = true
		go m.runLoop()
		go func() {
			_ = m.server.ListenAndServe()
		}()
	}
	m.updates <- fragment
	return nil
}

func (m *Manager) Stop() error {
	if !m.started || m.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := m.server.Shutdown(ctx)

	done := make(chan struct{})
	m.stopLoop <- done
	<-done

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
	lastFragment := ""

	for {
		select {
		case fragment := <-m.updates:
			lastFragment = fragment
			if conn != nil {
				if err := conn.WriteJSON(renderMessage{
					Type: "render",
					HTML: fragment,
				}); err != nil {
					_ = conn.Close()
					conn = nil
				}
			}

		case c := <-m.register:
			if conn != nil {
				_ = conn.Close()
			}
			conn = c

			if err := conn.WriteJSON(renderMessage{
				Type: "render",
				HTML: lastFragment,
			}); err != nil {
				_ = conn.Close()
				conn = nil
			}

		case c := <-m.unregister:
			if conn == c {
				_ = conn.Close()
				conn = nil
			}

		case done := <-m.stopLoop:
			if conn != nil {
				_ = conn.Close()
				conn = nil
			}
			close(done)
			return
		}
	}
}
