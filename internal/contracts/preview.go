package contracts

const (
	MessageTypeRender = "render"
	MessageTypeCursor = "cursor"
)

type RenderMessage struct {
	Type string `json:"type"`
	HTML string `json:"html"`
	Rev  uint64 `json:"rev"`
}

type CursorMessage struct {
	Type string `json:"type"`
	Line int    `json:"line"`
	Col  int    `json:"col"`
	Rev  uint64 `json:"rev"`
}
