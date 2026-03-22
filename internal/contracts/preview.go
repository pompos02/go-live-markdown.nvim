// Package contracts defines shared message types for host-browser communication.
package contracts

const (
	// MessageTypeRender updates the browser with rendered markdown HTML.
	MessageTypeRender = "render"
	// MessageTypeCursor updates the browser cursor/scroll position.
	MessageTypeCursor = "cursor"
	// MessageTypeGoToLine asks Neovim to move its cursor to a source line.
	MessageTypeGoToLine = "go_to_line"
	// MessageTypeToggleCheckbox asks Neovim to toggle a markdown task checkbox.
	MessageTypeToggleCheckbox = "toggle_checkbox"
)

// IncomingMessage is the minimal envelope used to route browser messages.
type IncomingMessage struct {
	Type string `json:"type"`
}

// GoToLineMessage requests a cursor jump in the editor.
type GoToLineMessage struct {
	Type string `json:"type"`
	Line int    `json:"line"`
}

// ToggleCheckboxMessage requests a task-list checkbox toggle in the editor.
type ToggleCheckboxMessage struct {
	Type string `json:"type"`
	Line int    `json:"line"`
	Rev  uint64 `json:"rev"`
}

// TOCItem represents a single table-of-contents heading entry.
type TOCItem struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Level int    `json:"level"`
	Line  int    `json:"line"`
}

// RenderMessage carries rendered HTML and revision metadata to the browser.
type RenderMessage struct {
	Type     string    `json:"type"`
	HTML     string    `json:"html"`
	TOC      []TOCItem `json:"toc"`
	Filename string    `json:"filename"`
	Rev      uint64    `json:"rev"`
}

// CursorMessage carries cursor position and revision metadata to the browser.
type CursorMessage struct {
	Type string `json:"type"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
	Rev  uint64 `json:"rev"`
}
