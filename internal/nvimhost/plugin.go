package nvimhost

import (
	"bytes"
	"fmt"

	"go-live-markdown/internal/markdown"
	"go-live-markdown/internal/preview"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

type App struct {
	renderer     *markdown.Renderer
	preview      *preview.Manager
	active       bool
	activeBuffer nvim.Buffer
}

func NewApp() *App {
	renderer := markdown.NewRenderer()
	return &App{
		renderer: renderer,
		preview:  preview.NewManager("127.0.0.1:7777", renderer.RenderShell()),
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
	p.HandleCommand(&plugin.CommandOptions{
		Name: "GoLiveMarkdownUpdate",
	}, app.GoLiveMarkdownUpdate)
	return nil
}

func (a *App) GoLiveMarkdownStart(v *nvim.Nvim) error {
	buf, err := v.CurrentBuffer()
	if err != nil {
		return err
	}
	a.active = true
	a.activeBuffer = buf
	if err := a.publishBuffer(v, buf); err != nil {
		return err
	}
	return v.Command(fmt.Sprintf(`echom "[go-live-markdown] preview: %s"`, a.preview.URL()))
}

func (a *App) GoLiveMarkdownUpdate(v *nvim.Nvim) error {
	if !a.active {
		return nil
	}

	buf, err := v.CurrentBuffer()
	if err != nil {
		return err
	}

	// here we could also query if the buffer is markdown
	if buf != a.activeBuffer {
		return nil
	}

	return a.publishBuffer(v, buf)
}

func (a *App) publishBuffer(v *nvim.Nvim, buf nvim.Buffer) error {
	lines, err := v.BufferLines(buf, 0, -1, true)
	if err != nil {
		return err
	}
	source := bytes.Join(lines, []byte("\n"))
	fragment, err := a.renderer.ConvertFragment(source)
	if err != nil {
		return err
	}

	return a.preview.StartOrUpdate(fragment)
}
