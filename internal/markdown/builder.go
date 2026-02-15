package markdown

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// BuildMarkdownPrint parses a markdown file and generates ESC/POS output
func BuildMarkdownPrint(filePath string, imageThreshold int) ([]byte, error) {
	// Read markdown file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Get directory of markdown file for resolving relative image paths
	markdownDir := filepath.Dir(filePath)

	// Create Goldmark parser with extensions
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,         // GitHub Flavored Markdown
			extension.TaskList,    // Task lists with [ ] and [x]
			NewBarcodeExtension(), // Custom barcode extension
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Auto-generate heading IDs
			parser.WithAttribute(),     // Enable attributes (helps with HTML parsing)
		),
	)

	// Parse markdown to AST
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	// Create ESC/POS renderer
	renderer := NewESCPOSRenderer(markdownDir, imageThreshold)

	// Render to ESC/POS
	var buf bytes.Buffer
	writer := &bufWriter{buf: &buf}
	if err := renderer.Render(writer, source, doc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// bufWriter wraps bytes.Buffer to implement util.BufWriter
type bufWriter struct {
	buf *bytes.Buffer
}

func (w *bufWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *bufWriter) WriteByte(c byte) error {
	return w.buf.WriteByte(c)
}

func (w *bufWriter) WriteRune(r rune) (int, error) {
	return w.buf.WriteRune(r)
}

func (w *bufWriter) WriteString(s string) (int, error) {
	return w.buf.WriteString(s)
}

func (w *bufWriter) Bytes() []byte {
	return w.buf.Bytes()
}

func (w *bufWriter) Available() int {
	return w.buf.Cap() - w.buf.Len()
}

func (w *bufWriter) Buffered() int {
	return w.buf.Len()
}

func (w *bufWriter) Flush() error {
	return nil
}
