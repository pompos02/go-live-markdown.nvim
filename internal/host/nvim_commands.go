package host

import (
	"bytes"
	"fmt"

	"go-live-markdown/internal/app"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

type Commands struct {
	preview      *app.LivePreview
	active       bool
	activeBuffer nvim.Buffer
}

func NewCommands() *Commands {
	return &Commands{preview: app.NewLivePreview("127.0.0.1:7777")}
}

func Register(p *plugin.Plugin) error {
	commands := NewCommands()
	p.Handle("poll", func() (string, error) {
		return "ok", nil
	})
	p.HandleCommand(&plugin.CommandOptions{
		Name: "GoLiveMarkdownStart",
	}, commands.GoLiveMarkdownStart)
	p.HandleCommand(&plugin.CommandOptions{
		Name: "GoLiveMarkdownUpdate",
	}, commands.GoLiveMarkdownUpdate)
	return nil
}

func (c *Commands) GoLiveMarkdownStart(v *nvim.Nvim) error {
	buf, err := v.CurrentBuffer()
	if err != nil {
		return err
	}
	c.active = true
	c.activeBuffer = buf
	if err := c.publishBuffer(v, buf); err != nil {
		return err
	}
	return v.Command(fmt.Sprintf(`echom "[go-live-markdown] preview: %s"`, c.preview.URL()))
}

func (c *Commands) GoLiveMarkdownUpdate(v *nvim.Nvim) error {
	if !c.active {
		return nil
	}

	buf, err := v.CurrentBuffer()
	if err != nil {
		return err
	}

	if buf != c.activeBuffer {
		return nil
	}

	return c.publishBuffer(v, buf)
}

func (c *Commands) publishBuffer(v *nvim.Nvim, buf nvim.Buffer) error {
	lines, err := v.BufferLines(buf, 0, -1, true)
	if err != nil {
		return err
	}
	source := bytes.Join(lines, []byte("\n"))
	return c.preview.PublishSource(source)
}
