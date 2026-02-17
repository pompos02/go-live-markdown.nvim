package nvimhost

import (
	"bytes"
	"fmt"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
	"go-live-markdown/internal/markdown"
	"go-live-markdown/internal/preview"
)

type App struct {
	renderer *markdown.Renderer
	preview  *preview.Manager
}

func NewApp() *App {
	return &App{
		renderer: markdown.NewRenderer(),
		preview:  preview.NewManager("127.0.0.1:7777"),
	}
}

func Register(p *plugin.Plugin) error {
	app := NewApp()
	p.Handle("poll", func() (string, error) {
		return "ok", nil
	})
	p.HandleCommand(&plugin.CommandOptions{
		Name: "GoLiveMarkdownStart",
	}, app.GoLiveMarkdownStart)
	return nil
}

func (a *App) GoLiveMarkdownStart(v *nvim.Nvim) error {
	buf, err := v.CurrentBuffer()
	if err != nil {
		return err
	}
	lines, err := v.BufferLines(buf, 0, -1, true)
	if err != nil {
		return err
	}
	source := bytes.Join(lines, []byte("\n"))
	page, err := a.renderer.RenderPage(source)
	if err != nil {
		return err
	}
	if err := a.preview.StartOrUpdate(page); err != nil {
		return err
	}
	return v.Command(fmt.Sprintf(`echom "[go-live-markdown] preview: %s"`, a.preview.URL()))
}
