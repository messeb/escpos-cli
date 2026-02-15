package markdown

import (
	"bytes"
	"strings"
	"testing"

	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/ast"
	text2 "github.com/yuin/goldmark/text"
)

func TestNewTableRenderer(t *testing.T) {
	tr := NewTableRenderer()

	if tr == nil {
		t.Fatal("NewTableRenderer() returned nil")
	}

	if tr.rows == nil {
		t.Error("rows slice is nil")
	}

	if tr.maxColumns != 0 {
		t.Errorf("maxColumns = %d, want 0", tr.maxColumns)
	}
}

func TestTableRenderer_calculateColumnWidths(t *testing.T) {
	tests := []struct {
		name       string
		maxColumns int
		wantTotal  int
	}{
		{
			name:       "2 columns",
			maxColumns: 2,
			wantTotal:  2,
		},
		{
			name:       "3 columns",
			maxColumns: 3,
			wantTotal:  3,
		},
		{
			name:       "5 columns",
			maxColumns: 5,
			wantTotal:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TableRenderer{
				rows:       [][]string{{"test"}},
				maxColumns: tt.maxColumns,
			}

			widths := tr.calculateColumnWidths()

			if len(widths) != tt.wantTotal {
				t.Errorf("calculateColumnWidths() returned %d widths, want %d", len(widths), tt.wantTotal)
			}

			// Calculate total width including borders
			totalWidth := 0
			for _, w := range widths {
				totalWidth += w
			}

			// Add border space: | col | col | = numCols*3 + 1
			borderWidth := tt.maxColumns*3 + 1
			totalWithBorders := totalWidth + borderWidth

			if totalWithBorders > maxTableWidth {
				t.Errorf("Total table width %d exceeds max %d", totalWithBorders, maxTableWidth)
			}
		})
	}
}

func TestTableRenderer_renderRow(t *testing.T) {
	tests := []struct {
		name       string
		row        []string
		colWidths  []int
		isHeader   bool
		wantStart  string
		wantEnd    string
	}{
		{
			name:      "simple 2-column row",
			row:       []string{"Item", "Price"},
			colWidths: []int{10, 10},
			isHeader:  true,
			wantStart: "|",
			wantEnd:   "|",
		},
		{
			name:      "3-column data row",
			row:       []string{"Coffee", "2", "9.00"},
			colWidths: []int{10, 5, 8},
			isHeader:  false,
			wantStart: "|",
			wantEnd:   "|",
		},
		{
			name:      "row with empty cells",
			row:       []string{"A", ""},
			colWidths: []int{10, 10},
			isHeader:  false,
			wantStart: "|",
			wantEnd:   "|",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TableRenderer{
				maxColumns: len(tt.colWidths),
				alignments: make([]extast.Alignment, len(tt.colWidths)),
			}

			result := tr.renderRow(tt.row, tt.colWidths, tt.isHeader)

			if !strings.HasPrefix(result, tt.wantStart) {
				t.Errorf("renderRow() starts with %q, want %q", result[:1], tt.wantStart)
			}

			if !strings.HasSuffix(result, tt.wantEnd) {
				t.Errorf("renderRow() ends with %q, want %q", result[len(result)-1:], tt.wantEnd)
			}

			// Count pipes to verify column count
			pipeCount := strings.Count(result, "|")
			expectedPipes := len(tt.colWidths) + 1
			if pipeCount != expectedPipes {
				t.Errorf("renderRow() has %d pipes, want %d", pipeCount, expectedPipes)
			}
		})
	}
}

func TestTableRenderer_renderTopBorder(t *testing.T) {
	tests := []struct {
		name      string
		colWidths []int
		wantStart string
		wantEnd   string
	}{
		{
			name:      "2 columns",
			colWidths: []int{10, 10},
			wantStart: "+",
			wantEnd:   "+",
		},
		{
			name:      "3 columns",
			colWidths: []int{8, 6, 8},
			wantStart: "+",
			wantEnd:   "+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TableRenderer{}
			result := tr.renderTopBorder(tt.colWidths)

			if !strings.HasPrefix(result, tt.wantStart) {
				t.Errorf("renderTopBorder() starts with %q, want %q", result[:1], tt.wantStart)
			}

			if !strings.HasSuffix(result, tt.wantEnd) {
				t.Errorf("renderTopBorder() ends with %q, want %q", result[len(result)-1:], tt.wantEnd)
			}

			// Should contain only +, -, and spaces
			for _, ch := range result {
				if ch != '+' && ch != '-' {
					t.Errorf("renderTopBorder() contains unexpected character %q", ch)
					break
				}
			}
		})
	}
}

func TestTableRenderer_renderReceiptFormat(t *testing.T) {
	tests := []struct {
		name     string
		rows     [][]string
		wantRows int
	}{
		{
			name: "simple receipt",
			rows: [][]string{
				{"Item", "Price"},
				{"Coffee", "4.50"},
				{"Tea", "3.00"},
			},
			wantRows: 3,
		},
		{
			name:     "empty table",
			rows:     [][]string{},
			wantRows: 0,
		},
		{
			name: "single row",
			rows: [][]string{
				{"Item", "Price"},
			},
			wantRows: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TableRenderer{
				rows:       tt.rows,
				maxColumns: 2,
				alignments: []extast.Alignment{extast.AlignLeft, extast.AlignRight},
			}

			result := tr.renderReceiptFormat()

			if tt.wantRows == 0 {
				if len(result) != 0 {
					t.Error("renderReceiptFormat() should return empty bytes for empty table")
				}
				return
			}

			// Check for presence of data from rows
			for i, row := range tt.rows {
				if i == 0 && len(tt.rows) > 1 {
					// Header row - should have solid line after it
					continue
				}
				for _, cell := range row {
					if !bytes.Contains(result, []byte(cell)) {
						t.Errorf("renderReceiptFormat() missing cell data: %s", cell)
					}
				}
			}
		})
	}
}

func TestTableRenderer_renderVerticalLayout(t *testing.T) {
	tests := []struct {
		name      string
		rows      [][]string
		wantEntry int
	}{
		{
			name: "wide table",
			rows: [][]string{
				{"Col1", "Col2", "Col3", "Col4"},
				{"A", "B", "C", "D"},
				{"E", "F", "G", "H"},
			},
			wantEntry: 2,
		},
		{
			name: "single data row",
			rows: [][]string{
				{"Name", "Age", "City", "Country"},
				{"John", "30", "NYC", "USA"},
			},
			wantEntry: 1,
		},
		{
			name:      "only headers (no data rows)",
			rows:      [][]string{{"A", "B", "C", "D"}},
			wantEntry: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TableRenderer{
				rows:       tt.rows,
				maxColumns: len(tt.rows[0]),
			}

			result := tr.renderVerticalLayout()

			// Count "Entry" markers
			entryCount := bytes.Count(result, []byte("--- Entry"))
			if entryCount != tt.wantEntry {
				t.Errorf("renderVerticalLayout() has %d entries, want %d", entryCount, tt.wantEntry)
			}

			// Check for header names in output (only if there are data rows)
			if len(tt.rows) > 1 {
				for _, header := range tt.rows[0] {
					if !bytes.Contains(result, []byte(header+":")) {
						t.Errorf("renderVerticalLayout() missing header: %s", header)
					}
				}
			}
		})
	}
}

func TestTableRenderer_Render(t *testing.T) {
	tests := []struct {
		name       string
		rows       [][]string
		maxColumns int
		wantFormat string
	}{
		{
			name: "2-column receipt format",
			rows: [][]string{
				{"Item", "Price"},
				{"Coffee", "4.50"},
			},
			maxColumns: 2,
			wantFormat: "receipt",
		},
		{
			name: "3-column bordered table",
			rows: [][]string{
				{"A", "B", "C"},
				{"1", "2", "3"},
			},
			maxColumns: 3,
			wantFormat: "bordered",
		},
		{
			name: "4-column vertical layout",
			rows: [][]string{
				{"A", "B", "C", "D"},
				{"1", "2", "3", "4"},
			},
			maxColumns: 4,
			wantFormat: "vertical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TableRenderer{
				rows:       tt.rows,
				maxColumns: tt.maxColumns,
				alignments: make([]extast.Alignment, tt.maxColumns),
			}

			result := tr.Render()

			if len(result) == 0 {
				t.Error("Render() returned empty output")
			}

			switch tt.wantFormat {
			case "receipt":
				// Receipt format should not have borders
				if bytes.Contains(result, []byte("+")) {
					t.Error("Receipt format should not contain border characters")
				}
			case "bordered":
				// Bordered format should have + and |
				if !bytes.Contains(result, []byte("+")) {
					t.Error("Bordered format should contain + characters")
				}
				if !bytes.Contains(result, []byte("|")) {
					t.Error("Bordered format should contain | characters")
				}
			case "vertical":
				// Vertical format should have "Entry" markers
				if !bytes.Contains(result, []byte("Entry")) {
					t.Error("Vertical format should contain Entry markers")
				}
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		text  string
		width int
		want  string
	}{
		{"test", 10, "test      "},
		{"hello", 5, "hello"},
		{"toolong", 4, "toolong"},
		{"", 5, "     "},
	}

	for _, tt := range tests {
		got := padRight(tt.text, tt.width)
		if got != tt.want {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.text, tt.width, got, tt.want)
		}
	}
}

func TestPadLeft(t *testing.T) {
	tests := []struct {
		text  string
		width int
		want  string
	}{
		{"test", 10, "      test"},
		{"hello", 5, "hello"},
		{"toolong", 4, "toolong"},
		{"", 5, "     "},
	}

	for _, tt := range tests {
		got := padLeft(tt.text, tt.width)
		if got != tt.want {
			t.Errorf("padLeft(%q, %d) = %q, want %q", tt.text, tt.width, got, tt.want)
		}
	}
}

func TestCenterText(t *testing.T) {
	tests := []struct {
		text  string
		width int
		want  string
	}{
		{"test", 10, "   test   "},
		{"hello", 10, "  hello   "},
		{"hi", 5, " hi  "},
		{"toolong", 4, "toolong"},
		{"", 5, "     "},
	}

	for _, tt := range tests {
		got := centerText(tt.text, tt.width)
		if got != tt.want {
			t.Errorf("centerText(%q, %d) = %q, want %q", tt.text, tt.width, got, tt.want)
		}
		if len(got) != len(tt.want) {
			t.Errorf("centerText(%q, %d) length = %d, want %d", tt.text, tt.width, len(got), len(tt.want))
		}
	}
}

func TestExtractCellText(t *testing.T) {
	// Create a simple table cell with text
	cell := extast.NewTableCell()
	text := ast.NewText()
	text.Segment = text2.NewSegment(0, 4)
	cell.AppendChild(cell, text)

	source := []byte("test")
	result := extractCellText(cell, source)

	if result != "test" {
		t.Errorf("extractCellText() = %q, want %q", result, "test")
	}
}

func TestExtractCellText_WithString(t *testing.T) {
	cell := extast.NewTableCell()
	str := ast.NewString([]byte("hello"))
	cell.AppendChild(cell, str)

	source := []byte("")
	result := extractCellText(cell, source)

	if result != "hello" {
		t.Errorf("extractCellText() = %q, want %q", result, "hello")
	}
}

func TestExtractCellText_Empty(t *testing.T) {
	cell := extast.NewTableCell()
	source := []byte("")
	result := extractCellText(cell, source)

	if result != "" {
		t.Errorf("extractCellText() = %q, want empty string", result)
	}
}

func TestExtractTextRecursive(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() (ast.Node, []byte)
		want   string
	}{
		{
			name: "nested text nodes",
			setup: func() (ast.Node, []byte) {
				parent := ast.NewParagraph()
				child := ast.NewText()
				child.Segment = text2.NewSegment(0, 5)
				parent.AppendChild(parent, child)
				return parent, []byte("hello")
			},
			want: "hello",
		},
		{
			name: "nested string nodes",
			setup: func() (ast.Node, []byte) {
				parent := ast.NewParagraph()
				child := ast.NewString([]byte("world"))
				parent.AppendChild(parent, child)
				return parent, []byte("")
			},
			want: "world",
		},
		{
			name: "deeply nested nodes",
			setup: func() (ast.Node, []byte) {
				parent := ast.NewParagraph()
				emph := ast.NewEmphasis(1)
				text := ast.NewText()
				text.Segment = text2.NewSegment(0, 4)
				emph.AppendChild(emph, text)
				parent.AppendChild(parent, emph)
				return parent, []byte("deep")
			},
			want: "deep",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, source := tt.setup()
			var buf bytes.Buffer
			extractTextRecursive(node, source, &buf)

			got := buf.String()
			if got != tt.want {
				t.Errorf("extractTextRecursive() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractCellText_NestedNodes(t *testing.T) {
	// Create a table cell with nested emphasis
	cell := extast.NewTableCell()
	emph := ast.NewEmphasis(2) // Bold
	text := ast.NewText()
	text.Segment = text2.NewSegment(0, 4)
	emph.AppendChild(emph, text)
	cell.AppendChild(cell, emph)

	source := []byte("bold")
	result := extractCellText(cell, source)

	if result != "bold" {
		t.Errorf("extractCellText() with nested nodes = %q, want %q", result, "bold")
	}
}
