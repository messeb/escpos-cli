package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// BarcodeType represents the type of barcode
type BarcodeType int

const (
	BarcodeQR BarcodeType = iota
	BarcodePDF417
	BarcodeDataMatrix
	BarcodeEAN13
)

// BarcodeBlock is an AST node for barcode code blocks
type BarcodeBlock struct {
	ast.BaseBlock
	BarcodeType BarcodeType
	Data        string
}

// Dump implements Node.Dump
func (n *BarcodeBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// Kind implements Node.Kind
func (n *BarcodeBlock) Kind() ast.NodeKind {
	return ast.NewNodeKind("BarcodeBlock")
}

// NewBarcodeBlock creates a new BarcodeBlock node
func NewBarcodeBlock(barcodeType BarcodeType, data string) *BarcodeBlock {
	return &BarcodeBlock{
		BarcodeType: barcodeType,
		Data:        data,
	}
}

// BarcodeParser parses barcode code blocks
type BarcodeParser struct{}

// NewBarcodeParser creates a new BarcodeParser
func NewBarcodeParser() parser.BlockParser {
	return &BarcodeParser{}
}

// Trigger returns characters that trigger this parser
func (b *BarcodeParser) Trigger() []byte {
	return []byte{'`'}
}

// Open tries to open a new barcode block
func (b *BarcodeParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()

	// Check for code fence (```)
	if len(line) < 3 || line[0] != '`' || line[1] != '`' || line[2] != '`' {
		return nil, parser.NoChildren
	}

	// Extract language identifier
	langStart := 3
	for langStart < len(line) && line[langStart] == ' ' {
		langStart++
	}

	langEnd := langStart
	for langEnd < len(line) && line[langEnd] != ' ' && line[langEnd] != '\n' {
		langEnd++
	}

	if langStart >= langEnd {
		return nil, parser.NoChildren
	}

	lang := string(line[langStart:langEnd])

	// Check if it's a barcode type
	var barcodeType BarcodeType
	var isBarcode bool

	switch lang {
	case "qr":
		barcodeType = BarcodeQR
		isBarcode = true
	case "pdf417":
		barcodeType = BarcodePDF417
		isBarcode = true
	case "datamatrix":
		barcodeType = BarcodeDataMatrix
		isBarcode = true
	case "ean13":
		barcodeType = BarcodeEAN13
		isBarcode = true
	default:
		return nil, parser.NoChildren
	}

	if !isBarcode {
		return nil, parser.NoChildren
	}

	// Consume the opening line
	reader.Advance(len(line))

	// Read the barcode data until closing ```
	var data []byte
	for {
		line, segment := reader.PeekLine()
		if segment.IsEmpty() {
			break
		}

		// Check for closing fence
		if len(line) >= 3 && line[0] == '`' && line[1] == '`' && line[2] == '`' {
			reader.Advance(segment.Len())
			break
		}

		// Append data
		data = append(data, line...)
		reader.Advance(segment.Len())
	}

	// Trim trailing newline if present
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	return NewBarcodeBlock(barcodeType, string(data)), parser.NoChildren
}

// Continue continues parsing
func (b *BarcodeParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

// Close finalizes the block
func (b *BarcodeParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// Nothing to do
}

// CanInterruptParagraph returns true if this parser can interrupt a paragraph
func (b *BarcodeParser) CanInterruptParagraph() bool {
	return true
}

// CanAcceptIndentedLine returns true if this parser can accept indented lines
func (b *BarcodeParser) CanAcceptIndentedLine() bool {
	return false
}

// BarcodeExtension is a Goldmark extension for barcode blocks
type BarcodeExtension struct{}

// Extend extends the Goldmark parser
func (e *BarcodeExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewBarcodeParser(), 100),
		),
	)
}

// NewBarcodeExtension creates a new BarcodeExtension
func NewBarcodeExtension() goldmark.Extender {
	return &BarcodeExtension{}
}
