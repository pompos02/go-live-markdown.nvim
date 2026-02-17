/* Package preview is responsilbe for the server implementation of the plugin */
package preview

import (
	"context"
	"net/http"
	"time"
)

type Manager struct {
	addr string

	html    string
	started bool
	server  *http.Server
}

func NewManager(addr string) *Manager {
	return &Manager{addr: addr}
}

func (m *Manager) URL() string {
	return "http://" + m.addr
}

func (m *Manager) StartOrUpdate(html string) error {
	m.html = html

	if m.started {
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(m.html))
	})

	m.server = &http.Server{
		Addr:    m.addr,
		Handler: mux,
	}

	m.started = true

	go func() {
		_ = m.server.ListenAndServe()
	}()

	return nil
}

func (m *Manager) Stop() error {
	if !m.started || m.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := m.server.Shutdown(ctx)
	m.started = false
	m.server = nil
	return err
}
