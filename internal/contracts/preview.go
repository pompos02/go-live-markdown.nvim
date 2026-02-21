package contracts

const (
	// MessageTypeRender updates the browser with rendered markdown HTML.
	MessageTypeRender = "render"
	// MessageTypeCursor updates the browser cursor/scroll position.
	MessageTypeCursor = "cursor"
	// MessageTypeGoToLine asks Neovim to move its cursor to a source line.
	MessageTypeGoToLine = "go_to_line"
)

// IncomingMessage is the minimal envelope used to route browser messages.
type IncomingMessage struct {
	Type string
}

// GoToLineMessage requests a cursor jump in the editor.
type GoToLineMessage struct {
	Type string `json:"type"`
	Line int    `json:"line"`
}

// RenderMessage carries rendered HTML and revision metadata to the browser.
type RenderMessage struct {
	Type     string `json:"type"`
	HTML     string `json:"html"`
	Filename string `json:"filename"`
	Rev      uint64 `json:"rev"`
}

// CursorMessage carries cursor position and revision metadata to the browser.
type CursorMessage struct {
	Type string `json:"type"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
	Rev  uint64 `json:"rev"`
}
