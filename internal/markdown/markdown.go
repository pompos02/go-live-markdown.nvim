package markdown

import (
	"bytes"
	_ "embed"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	alertcallouts "github.com/zmtcreative/gm-alert-callouts"
	"go.abhg.dev/goldmark/mermaid"
)

type Renderer struct {
	md goldmark.Markdown
}

//go:embed page.html
var pageTemplate string

func NewRenderer() *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(
			alertcallouts.NewAlertCallouts( // Alerts !NOTES, !IMPORTANT e.t.c
				alertcallouts.UseGFMStrictIcons(),
				alertcallouts.WithFolding(true),
			),
			&mermaid.Extender{},
			extension.GFM,                  // GitHub Flavored Markdown
			extension.Table,                // Tables
			extension.Strikethrough,        // ~~text~~
			extension.TaskList,             // - [x] tasks
			extension.Linkify,              // auto links
			highlighting.NewHighlighting(), // Syntax highlighting
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Auto IDs for headings
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // Allow raw HTML inside markdown
		),
	)
	return &Renderer{md: md}
}

func (r *Renderer) ConvertFragment(source []byte) (string, error) {
	var buf bytes.Buffer
	if err := r.md.Convert(source, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (r *Renderer) RenderPage(source []byte) (string, error) {
	fragment, err := r.ConvertFragment(source)
	if err != nil {
		return "", err
	}
	return strings.Replace(pageTemplate, "{{CONTENT}}", fragment, 1), nil
}

func (r *Renderer) RenderShell() string {
	return strings.Replace(pageTemplate, "{{CONTENT}}", "", 1)
}
