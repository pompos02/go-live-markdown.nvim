package render

import (
	"bytes"
	_ "embed"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extensionast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	alertcallouts "github.com/zmtcreative/gm-alert-callouts"
)

const mdLineAttribute = "data-md-line"

// Renderer is a wrapper around the Goldmark mardown parser with pre-configured extensions
type Renderer struct {
	md goldmark.Markdown
}

//go:embed page.html
var pageTemplate string

func NewRenderer() *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(
			alertcallouts.AlertCallouts,
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
			extension.Linkify,
			highlighting.NewHighlighting(
				highlighting.WithWrapperRenderer(renderHighlightedCodeWrapper),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			renderer.WithNodeRenderers(
				util.Prioritized(newCodeBlockRenderer(), 300),
			),
		),
	)
	return &Renderer{md: md}
}

func (r *Renderer) ConvertFragment(source []byte) (string, error) {
	doc := r.md.Parser().Parse(text.NewReader(source))
	annotateBlockSourceLines(doc, source)

	var buf bytes.Buffer
	if err := r.md.Renderer().Render(&buf, source, doc); err != nil {
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

func annotateBlockSourceLines(doc ast.Node, source []byte) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || !shouldAnnotateNode(n) {
			return ast.WalkContinue, nil
		}

		offset, ok := firstNodeOffset(n)
		if !ok {
			return ast.WalkContinue, nil
		}

		n.SetAttributeString(mdLineAttribute, strconv.Itoa(offsetToLine(source, offset)))
		return ast.WalkContinue, nil
	})
}

func shouldAnnotateNode(n ast.Node) bool {
	switch n.Kind() {
	case ast.KindHeading,
		ast.KindParagraph,
		ast.KindBlockquote,
		ast.KindCodeBlock,
		ast.KindFencedCodeBlock,
		ast.KindList,
		ast.KindListItem,
		ast.KindThematicBreak,
		extensionast.KindTable:
		return true
	default:
		return false
	}
}

func firstNodeOffset(n ast.Node) (int, bool) {
	if n == nil {
		return 0, false
	}

	if lines := n.Lines(); lines != nil && lines.Len() > 0 {
		return lines.At(0).Start, true
	}

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if offset, ok := firstNodeOffset(child); ok {
			return offset, true
		}
	}

	return 0, false
}

func offsetToLine(source []byte, offset int) int {
	if offset < 0 {
		offset = 0
	}

	if offset > len(source) {
		offset = len(source)
	}

	return bytes.Count(source[:offset], []byte{'\n'}) + 1
}

type codeBlockRenderer struct{}

func newCodeBlockRenderer() renderer.NodeRenderer {
	return &codeBlockRenderer{}
}

func (r *codeBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
}

func (r *codeBlockRenderer) renderCodeBlock(
	w util.BufWriter,
	source []byte,
	n ast.Node,
	entering bool,
) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<pre")
		if n.Attributes() != nil {
			html.RenderAttributes(w, n, html.GlobalAttributeFilter)
		}
		_, _ = w.WriteString("><code>")

		lines := n.Lines()
		if lines != nil {
			for i := 0; i < lines.Len(); i++ {
				line := lines.At(i)
				html.DefaultWriter.RawWrite(w, line.Value(source))
			}
		}
	} else {
		_, _ = w.WriteString("</code></pre>\n")
	}

	return ast.WalkContinue, nil
}

func renderHighlightedCodeWrapper(w util.BufWriter, context highlighting.CodeBlockContext, entering bool) {
	line, ok := highlightedCodeLine(context)
	if !ok {
		return
	}

	if entering {
		_, _ = w.WriteString("<div ")
		_, _ = w.WriteString(mdLineAttribute)
		_, _ = w.WriteString(`="`)
		_, _ = w.WriteString(line)
		_, _ = w.WriteString(`">`)
		return
	}

	_, _ = w.WriteString("</div>")
}

func highlightedCodeLine(context highlighting.CodeBlockContext) (string, bool) {
	if context == nil {
		return "", false
	}

	attrs := context.Attributes()
	if attrs == nil {
		return "", false
	}

	v, ok := attrs.GetString(mdLineAttribute)
	if !ok {
		return "", false
	}

	switch typed := v.(type) {
	case string:
		return typed, typed != ""
	case []byte:
		if len(typed) == 0 {
			return "", false
		}
		return string(typed), true
	default:
		return "", false
	}
}
