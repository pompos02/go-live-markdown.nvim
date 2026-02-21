package app

import (
	"go-live-markdown/internal/contracts"
	"go-live-markdown/internal/render"
	httpserver "go-live-markdown/internal/transport/http"
)

// LivePreview is a coordinator between markdown rendering and HTTP delivery.
type LivePreview struct {
	renderer *render.Renderer
	preview  *httpserver.PreviewServer
}

// NewLivePreview wires the markdown renderer with the HTTP preview transport.
func NewLivePreview(addr string) *LivePreview {
	renderer := render.NewRenderer()
	return &LivePreview{
		renderer: renderer,
		preview:  httpserver.NewPreviewServer(addr, renderer.RenderShell()),
	}
}

// URL returns the preview server URL that users can open in a browser.
func (s *LivePreview) URL() string {
	return s.preview.URL()
}

// PublishSource renders markdown source and publishes it to the preview server.
func (s *LivePreview) PublishSource(source []byte, path string) error {
	fragment, err := s.renderer.ConvertFragmentWithSourcePath(source, path)
	if err != nil {
		return err
	}

	return s.preview.StartOrUpdate(fragment, path)
}

// PublishCursor forwards the current editor cursor position to the browser.
func (s *LivePreview) PublishCursor(line int, col int) error {
	return s.preview.UpdateCursor(contracts.CursorMessage{
		Type: contracts.MessageTypeCursor,
		Line: line,
		Col:  col,
	})
}

// SetGoToLineHandler registers a callback for browser-initiated go-to-line events.
func (s *LivePreview) SetGoToLineHandler(fn func(contracts.GoToLineMessage)) {
	s.preview.SetGoToLineHandler(fn)
}
