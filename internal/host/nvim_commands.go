package host

import (
	"bytes"
	"fmt"

	"go-live-markdown/internal/app"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

// Commands is a state container for Neovim command handlers.
// It tracks the active buffer and delegates preview functionality
// to the LivePreview service.
type Commands struct {
	preview      *app.LivePreview
	active       bool
	activeBuffer nvim.Buffer

	lastCursorLine int
	lastCursorCol  int
}

func NewCommands() *Commands {
	return &Commands{preview: app.NewLivePreview("127.0.0.1:7777")}
}

// Register registers Neovim command/function handlers.
func Register(p *plugin.Plugin) error {
	commands := NewCommands()

	p.Handle("poll", func() (string, error) {
		return "ok", nil
	})

	p.HandleCommand(&plugin.CommandOptions{
		Name: "GoLiveMarkdownStart",
	}, commands.GoLiveMarkdownStart)

	p.HandleFunction(&plugin.FunctionOptions{
		Name: "GoLiveMarkdownInternalUpdate",
	}, commands.GoLiveMarkdownUpdate)

	p.HandleFunction(&plugin.FunctionOptions{
		Name: "GoLiveMarkdownInternalCursor",
	}, commands.GoLiveMarkdownCursor)

	return nil
}

func (c *Commands) GoLiveMarkdownStart(v *nvim.Nvim) error {
	buf, err := v.CurrentBuffer()
	if err != nil {
		return err
	}
	c.active = true
	c.activeBuffer = buf
	c.lastCursorLine = 0
	c.lastCursorCol = 0

	if err := c.publishBuffer(v, buf); err != nil {
		return err
	}

	if err := c.publishCursor(v); err != nil {
		return err
	}

	return v.Command(fmt.Sprintf(`echom "[go-live-markdown] preview: %s"`, c.preview.URL()))
}

func (c *Commands) GoLiveMarkdownUpdate(v *nvim.Nvim) error {
	if !c.active {
		return nil
	}

	buf, err := c.currentActiveBuffer(v)
	if err != nil {
		return err
	}

	return c.publishBuffer(v, buf)
}

func (c *Commands) GoLiveMarkdownCursor(v *nvim.Nvim) error {
	if !c.active {
		return nil
	}

	buf, err := c.currentActiveBuffer(v)
	if err != nil {
		return err
	}
	if buf == 0 {
		return nil
	}

	return c.publishCursor(v)
}

func (c *Commands) currentActiveBuffer(v *nvim.Nvim) (nvim.Buffer, error) {
	buf, err := v.CurrentBuffer()
	if err != nil {
		return 0, err
	}

	if buf != c.activeBuffer {
		return 0, nil
	}

	return buf, nil
}

func (c *Commands) publishBuffer(v *nvim.Nvim, buf nvim.Buffer) error {
	if buf == 0 {
		return nil
	}

	lines, err := v.BufferLines(buf, 0, -1, true)
	if err != nil {
		return err
	}

	source := bytes.Join(lines, []byte("\n"))
	return c.preview.PublishSource(source)
}

func (c *Commands) publishCursor(v *nvim.Nvim) error {
	var line int
	if err := v.Eval(`line(".")`, &line); err != nil {
		return err
	}

	var col int
	if err := v.Eval(`col(".")`, &col); err != nil {
		return err
	}

	if line == c.lastCursorLine && col == c.lastCursorCol {
		return nil
	}

	c.lastCursorLine = line
	c.lastCursorCol = col
	return c.preview.PublishCursor(line, col)
}
