package contracts

const (
	MessageTypeRender = "render"
	MessageTypeCursor = "cursor"
	MessageTypeGoToLine = "go_to_line"
)

type IncomingMessage struct {
	Type string
}

type GoToLineMessage struct {
	Type string `json:"type"`
	Line int    `json:"line"`
}

type RenderMessage struct {
	Type     string `json:"type"`
	HTML     string `json:"html"`
	Filename string `json:"filename"`
	Rev      uint64 `json:"rev"`
}

type CursorMessage struct {
	Type string `json:"type"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
	Rev  uint64 `json:"rev"`
}
