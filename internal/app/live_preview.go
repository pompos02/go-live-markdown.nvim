package app

import (
	"go-live-markdown/internal/render"
	httptransport "go-live-markdown/internal/transport/http"
)

// Coordinator between markdown rendering and HTTP delivery
type LivePreview struct {
	renderer *render.Renderer
	preview  *httptransport.Manager
}

func NewLivePreview(addr string) *LivePreview {
	renderer := render.NewRenderer()
	return &LivePreview{
		renderer: renderer,
		preview:  httptransport.NewManager(addr, renderer.RenderShell()),
	}
}

func (s *LivePreview) URL() string {
	return s.preview.URL()
}

func (s *LivePreview) PublishSource(source []byte) error {
	fragment, err := s.renderer.ConvertFragment(source)
	if err != nil {
		return err
	}

	return s.preview.StartOrUpdate(fragment)
}
