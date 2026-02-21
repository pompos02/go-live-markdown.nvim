package render

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"path/filepath"
	"strconv"
	"strings"

	chromahtml "github.com/alecthomas/chroma/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extensionast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
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
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
				),
			),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	return &Renderer{md: md}
}

// ConvertFragment parses markdown source and returns the HTML fragment
// with data-md-line attributes attached to block elements.
func (r *Renderer) ConvertFragment(source []byte) (string, error) {
	return r.ConvertFragmentWithSourcePath(source, "")
}

// ConvertFragmentWithSourcePath parses markdown source and returns the HTML
// fragment with data-md-line attributes attached to block elements.
//
// If sourcePath is set, local image destinations are rewritten to the
// preview asset path format expected by the HTTP layer.
func (r *Renderer) ConvertFragmentWithSourcePath(source []byte, sourcePath string) (string, error) {
	doc := r.md.Parser().Parse(text.NewReader(source))
	decorateAST(doc, source, sourcePath)

	var buf bytes.Buffer
	if err := r.md.Renderer().Render(&buf, source, doc); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderPage returns a complete HTML page with the markdown rendered inside.
// The fragment is inserted into the page template at the {{CONTENT}} placeholder.
func (r *Renderer) RenderPage(source []byte) (string, error) {
	fragment, err := r.ConvertFragment(source)
	if err != nil {
		return "", err
	}
	return strings.Replace(pageTemplate, "{{CONTENT}}", fragment, 1), nil
}

// RenderShell returns an empty HTML page shell for the initial WebSocket connection.
// Content will be injected dynamically via WebSocket messages.
func (r *Renderer) RenderShell() string {
	return strings.Replace(pageTemplate, "{{CONTENT}}", "", 1)
}

// decorateAST walks the AST once and applies render metadata.
// It attaches data-md-line to block-level elements for cursor sync and,
// when sourcePath is available, rewrites local image destinations to /@mdfs/.
func decorateAST(doc ast.Node, source []byte, sourcePath string) {
	baseDir := ""
	if sourcePath != "" {
		baseDir = filepath.Dir(sourcePath)
	}

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if shouldAnnotateNode(n) {
			offset, ok := firstNodeOffset(n)
			if ok {
				n.SetAttributeString(mdLineAttribute, strconv.Itoa(offsetToLine(source, offset)))
			}
		}

		img, ok := n.(*ast.Image)
		if !ok {
			return ast.WalkContinue, nil
		}

		rawDest := strings.TrimSpace(string(img.Destination))
		if rawDest == "" {
			return ast.WalkContinue, nil
		}

		lowerDest := strings.ToLower(rawDest)
		if strings.HasPrefix(lowerDest, "http://") ||
			strings.HasPrefix(lowerDest, "https://") ||
			strings.HasPrefix(lowerDest, "data:") ||
			strings.HasPrefix(lowerDest, "blob:") ||
			strings.HasPrefix(lowerDest, "file://") ||
			strings.HasPrefix(lowerDest, "//") ||
			strings.HasPrefix(lowerDest, "#") ||
			strings.HasPrefix(lowerDest, "/@mdfs/") {
			return ast.WalkContinue, nil
		}

		if filepath.IsAbs(rawDest) {
			img.Destination = []byte("/@mdfs/" + base64.RawURLEncoding.EncodeToString([]byte(filepath.Clean(rawDest))))
			img.SetAttributeString("loading", "lazy")
			img.SetAttributeString("decoding", "async")
			return ast.WalkContinue, nil
		}

		if baseDir == "" {
			return ast.WalkContinue, nil
		}

		resolved := filepath.Clean(filepath.Join(baseDir, rawDest))
		img.Destination = []byte("/@mdfs/" + base64.RawURLEncoding.EncodeToString([]byte(resolved)))
		img.SetAttributeString("loading", "lazy")
		img.SetAttributeString("decoding", "async")

		return ast.WalkContinue, nil
	})
}

// shouldAnnotateNode returns true for block-level element types that should
// receive line metadata. These are the elements that map directly to source lines.
func shouldAnnotateNode(n ast.Node) bool {
	switch n.Kind() {
	case ast.KindHeading,
		ast.KindParagraph,
		ast.KindBlockquote,
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

// firstNodeOffset returns the byte offset of the first line in a node.
// It first checks if the node has its own lines (most block elements do).
// If not, it recursively searches children to find the first meaningful offset.
// This handles nodes like lists that contain list items with actual content.
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

// offsetToLine converts a byte offset to a 1-based line number.
// It counts how many newline characters appear before the offset position
// The offset is clamped to the valid range [0, len(source)].
func offsetToLine(source []byte, offset int) int {
	if offset < 0 {
		offset = 0
	}

	if offset > len(source) {
		offset = len(source)
	}

	return bytes.Count(source[:offset], []byte{'\n'}) + 1
}

// renderHighlightedCodeWrapper wraps syntax-highlighted code blocks in a div
// with the data-md-line attribute. This is a custom wrapper renderer used by
// the goldmark-highlighting extension to preserve line metadata for code blocks.
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

// highlightedCodeLine extracts the line number attribute from a code block's
// rendering context. This attribute was set during the annotateBlockSourceLines
// walk and needs to be transferred to the wrapper div.
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
