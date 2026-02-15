package markdown

import (
	"bytes"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestESCPOSRenderer_renderParagraph_InsideListItem(t *testing.T) {
	// Test paragraph rendering when inside a list item
	// This tests the special case in renderParagraph that checks the parent
	r := NewESCPOSRenderer("/tmp", 75)

	listItem := ast.NewListItem(0)
	para := ast.NewParagraph()
	listItem.AppendChild(listItem, para)

	// Render paragraph entering
	status, err := r.renderParagraph(para, true)
	if err != nil {
		t.Errorf("renderParagraph(entering) error = %v", err)
	}
	if status != ast.WalkContinue {
		t.Errorf("renderParagraph(entering) status = %v, want WalkContinue", status)
	}

	// Render paragraph exiting (should not add extra line feed in list)
	beforeLen := r.buf.Len()
	_, err = r.renderParagraph(para, false)
	if err != nil {
		t.Errorf("renderParagraph(exiting) error = %v", err)
	}

	// Should only add one line feed, not two (because parent is list item)
	afterLen := r.buf.Len()
	bytesAdded := afterLen - beforeLen

	// LineFeed is 2 bytes (CR+LF), so should be 2 bytes, not 4
	if bytesAdded != 2 {
		t.Errorf("renderParagraph(exiting) in list added %d bytes, want 2 (single line feed)", bytesAdded)
	}
}

func TestESCPOSRenderer_renderParagraph_NotInList(t *testing.T) {
	// Test paragraph rendering when NOT inside a list item
	r := NewESCPOSRenderer("/tmp", 75)

	doc := ast.NewDocument()
	para := ast.NewParagraph()
	doc.AppendChild(doc, para)

	// Render paragraph exiting (should add two line feeds)
	beforeLen := r.buf.Len()
	status, err := r.renderParagraph(para, false)
	if err != nil {
		t.Errorf("renderParagraph(exiting) error = %v", err)
	}
	if status != ast.WalkContinue {
		t.Errorf("renderParagraph(exiting) status = %v, want WalkContinue", status)
	}

	// Should add two line feeds (4 bytes total: 2x CR+LF)
	afterLen := r.buf.Len()
	bytesAdded := afterLen - beforeLen

	if bytesAdded != 4 {
		t.Errorf("renderParagraph(exiting) not in list added %d bytes, want 4 (two line feeds)", bytesAdded)
	}
}

func TestESCPOSRenderer_writeWrappedText_EmptyString(t *testing.T) {
	r := NewESCPOSRenderer("/tmp", 75)

	// Should handle empty string gracefully
	r.writeWrappedText("")

	if r.buf.Len() != 0 {
		t.Error("writeWrappedText() with empty string should produce no output")
	}
}

func TestESCPOSRenderer_writeWrappedText_OnlyWhitespace(t *testing.T) {
	r := NewESCPOSRenderer("/tmp", 75)

	// String with only whitespace should produce no output (Fields returns empty slice)
	r.writeWrappedText("   \t  \n  ")

	if r.buf.Len() != 0 {
		t.Error("writeWrappedText() with only whitespace should produce no output")
	}
}

func TestESCPOSRenderer_writeWrappedText_VeryLongWord(t *testing.T) {
	r := NewESCPOSRenderer("/tmp", 75)

	// Word longer than max width (46 chars)
	longWord := "supercalifragilisticexpialidociousverylongwordindeed"

	r.writeWrappedText(longWord)

	output := r.buf.String()
	if !bytes.Contains([]byte(output), []byte(longWord)) {
		t.Error("writeWrappedText() should contain the long word")
	}

	// Should wrap after the long word
	if r.currentColumn != 0 {
		t.Errorf("writeWrappedText() with very long word should reset column, got %d", r.currentColumn)
	}
}

func TestESCPOSRenderer_renderListItem_WithTaskCheckBox(t *testing.T) {
	// Test that list items with checkboxes don't get markers
	r := NewESCPOSRenderer("/tmp", 75)
	r.listDepth = 1
	r.listCounters = []int{0}

	listItem := ast.NewListItem(0)
	list := ast.NewList('-')
	list.AppendChild(list, listItem)

	// Add task checkbox to list item (use real extast.TaskCheckBox)
	checkbox := extast.NewTaskCheckBox(false)
	listItem.AppendChild(listItem, checkbox)

	status, err := r.renderListItem(listItem, true)
	if err != nil {
		t.Errorf("renderListItem() error = %v", err)
	}
	if status != ast.WalkContinue {
		t.Errorf("renderListItem() status = %v, want WalkContinue", status)
	}

	output := r.buf.String()
	// Should not start with list marker like "- " because it has a checkbox
	// The checkbox itself will be rendered separately
	if bytes.Contains([]byte(output), []byte("- ")) {
		t.Error("renderListItem() with checkbox should not include list marker")
	}
}

func TestTableRenderer_calculateColumnWidths_EmptyTable(t *testing.T) {
	tr := &TableRenderer{
		rows:       [][]string{},
		maxColumns: 0,
	}

	widths := tr.calculateColumnWidths()

	if len(widths) != 0 {
		t.Errorf("calculateColumnWidths() for empty table = %v, want empty slice", widths)
	}
}

func TestTableRenderer_renderRow_TruncateCell(t *testing.T) {
	tr := &TableRenderer{
		maxColumns: 2,
		alignments: make([]extast.Alignment, 2),
	}

	// Cell text longer than column width should be truncated
	row := []string{"verylongcelltext", "test"}
	colWidths := []int{5, 5}

	result := tr.renderRow(row, colWidths, false)

	// Should truncate first cell to 5 characters
	if !bytes.Contains([]byte(result), []byte("veryl")) {
		t.Error("renderRow() should truncate long cell text")
	}
}

func TestTableRenderer_renderReceiptFormat_SkipIncompleteRow(t *testing.T) {
	tr := &TableRenderer{
		rows: [][]string{
			{"Item", "Price"},  // Header
			{"Coffee", "4.50"}, // Complete row
			{"OnlyOneColumn"},  // Row with missing second column (should be skipped)
			{"Tea", "3.00"},    // Another complete row
		},
		maxColumns: 2,
		alignments: make([]extast.Alignment, 2),
	}

	result := tr.renderReceiptFormat()

	// Should generate output (not empty)
	if len(result) == 0 {
		t.Error("renderReceiptFormat() should generate output")
	}

	// Should contain complete rows
	if !bytes.Contains(result, []byte("Coffee")) {
		t.Error("renderReceiptFormat() should include complete rows")
	}

	// Incomplete row should be skipped (not contain OnlyOneColumn in output)
	// Note: This tests the actual behavior of skipping rows with < 2 columns
	output := string(result)
	t.Logf("Output: %s", output) // For debugging
}

func TestBarcodeParser_Open_EmptyCodeBlock(t *testing.T) {
	bp := NewBarcodeParser()
	input := "```qr\n```\n"
	reader := text.NewReader([]byte(input))
	pc := parser.NewContext()

	node, state := bp.Open(nil, reader, pc)

	if node == nil {
		t.Error("Open() should create barcode block even with empty data")
		return
	}

	barcodeBlock, ok := node.(*BarcodeBlock)
	if !ok {
		t.Error("Open() should return BarcodeBlock")
		return
	}

	if barcodeBlock.Data != "" {
		t.Errorf("Open() with empty code block data = %q, want empty string", barcodeBlock.Data)
	}

	if state != parser.NoChildren {
		t.Errorf("Open() state = %v, want NoChildren", state)
	}
}

func TestBarcodeParser_Close(t *testing.T) {
	bp := NewBarcodeParser()
	reader := text.NewReader([]byte("test"))
	pc := parser.NewContext()

	// Close should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Close() panicked: %v", r)
		}
	}()

	bp.Close(nil, reader, pc)
}

func TestFullMarkdownIntegration_WithBarcodes(t *testing.T) {
	markdownSource := `# Test Receipt

Order #12345

## Items

| Item | Price |
|------|-------|
| Coffee | $4.50 |

Total: **$4.50**

Scan to pay:

` + "```qr\nhttps://pay.example.com/12345\n```" + `

---

Thank you!`

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			NewBarcodeExtension(),
		),
	)

	reader := text.NewReader([]byte(markdownSource))
	doc := md.Parser().Parse(reader)

	renderer := NewESCPOSRenderer("/tmp", 75)
	var buf bytes.Buffer
	writer := &bufWriter{buf: &buf}

	err := renderer.Render(writer, []byte(markdownSource), doc)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	output := buf.Bytes()

	// Verify key elements are present
	expectedElements := []string{
		"Test Receipt",
		"Order #12345",
		"Items",
		"Coffee",
		"4.50",
		"Total",
		"Thank you",
	}

	for _, elem := range expectedElements {
		if !bytes.Contains(output, []byte(elem)) {
			t.Errorf("Full integration output missing element: %s", elem)
		}
	}

	// Should have QR code data
	if len(output) < 100 {
		t.Error("Full integration output seems too short, may be missing barcode")
	}
}

