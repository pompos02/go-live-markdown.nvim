package host

import (
	"bytes"
	"fmt"

	"go-live-markdown/internal/app"
	"go-live-markdown/internal/contracts"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
)

// Commands is a state container for Neovim command handlers.
// It tracks the active buffer and delegates preview functionality
// to the LivePreview service.
type Commands struct {
	preview *app.LivePreview
	active  bool

	nv *nvim.Nvim

	lastCursorLine int
	lastCursorCol  int
}

func NewCommands() *Commands {
	preview := app.NewLivePreview("127.0.0.1:7777")
	c := &Commands{preview: preview}

	preview.SetGoToLineHandler(func(msg contracts.GoToLineMessage) {
		c.handleGoToLine(msg)
	})
	return c
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
	c.active = true
	c.lastCursorLine = 0
	c.lastCursorCol = 0
	c.nv = v

	if err := c.publishBuffer(v); err != nil {
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

	return c.publishBuffer(v)
}

func (c *Commands) GoLiveMarkdownCursor(v *nvim.Nvim) error {
	if !c.active {
		return nil
	}
	return c.publishCursor(v)
}

func (c *Commands) currentPath(v *nvim.Nvim) (string, error) {
	absPath, err := v.BufferName(0)
	if err != nil {
		return "", err
	}
	// filename := filepath.Base(name)

	return absPath, nil
}

func (c *Commands) publishBuffer(v *nvim.Nvim) error {
	buf, err := v.CurrentBuffer()
	if err != nil {
		return nil
	}

	lines, err := v.BufferLines(buf, 0, -1, true)
	if err != nil {
		return err
	}

	source := bytes.Join(lines, []byte("\n"))
	path, err := c.currentPath(v)
	if err != nil {
		return err
	}
	return c.preview.PublishSource(source, path)
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

func (c *Commands) handleGoToLine(msg contracts.GoToLineMessage) {
	if !c.active || c.nv == nil {
		return
	}

	v := c.nv

	line := msg.Line
	if line == c.lastCursorLine {
		return 
	}

	win, err := v.CurrentWindow()
	if err != nil {
		return
	}
	if err := v.SetWindowCursor(win, [2]int{line, 0}); err != nil {
		return
	}

	_ = v.Command("normal! zz")
	c.lastCursorLine = line
	c.lastCursorCol = 0

	
}
