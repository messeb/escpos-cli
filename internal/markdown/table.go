package markdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/messeb/escpos-cli/internal/escpos"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

const (
	maxTableWidth = 46 // POS-80 thermal printer width in characters (80mm)
)

// TableRenderer handles table rendering for thermal printers
type TableRenderer struct {
	rows       [][]string
	maxColumns int
	alignments []extast.Alignment // Column alignments
}

// NewTableRenderer creates a new TableRenderer
func NewTableRenderer() *TableRenderer {
	return &TableRenderer{
		rows: [][]string{},
	}
}

// ExtractTableData extracts table data from AST
func (tr *TableRenderer) ExtractTableData(table *extast.Table, source []byte) {
	tr.rows = [][]string{}
	tr.maxColumns = 0
	tr.alignments = []extast.Alignment{}

	// Walk through table rows (including header)
	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		// Process both TableHeader and TableRow
		if child.Kind() == extast.KindTableHeader || child.Kind() == extast.KindTableRow {
			rowData := []string{}

			// Extract cells and alignments (from first row only)
			colIndex := 0
			for cell := child.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if tableCell, ok := cell.(*extast.TableCell); ok {
					cellText := extractCellText(tableCell, source)
					rowData = append(rowData, cellText)

					// Store alignment from first row
					if len(tr.rows) == 0 {
						tr.alignments = append(tr.alignments, tableCell.Alignment)
					}
					colIndex++
				}
			}

			tr.rows = append(tr.rows, rowData)
			if len(rowData) > tr.maxColumns {
				tr.maxColumns = len(rowData)
			}
		}
	}
}

// Render renders the table to ESC/POS
func (tr *TableRenderer) Render() []byte {
	var buf bytes.Buffer

	// If more than 3 columns, use vertical layout
	if tr.maxColumns > 3 {
		return tr.renderVerticalLayout()
	}

	// For 2-column tables, use simple receipt format (no borders)
	if tr.maxColumns == 2 {
		return tr.renderReceiptFormat()
	}

	// Calculate column widths
	colWidths := tr.calculateColumnWidths()

	// Render table with borders (for 3-column tables)
	for i, row := range tr.rows {
		// Render separator before first row and after header
		switch i {
		case 0:
			buf.WriteString(tr.renderTopBorder(colWidths))
			buf.WriteByte('\n')
		case 1:
			buf.WriteString(tr.renderMiddleBorder(colWidths))
			buf.WriteByte('\n')
		}

		// Render row
		buf.WriteString(tr.renderRow(row, colWidths, i == 0))
		buf.WriteByte('\n')
	}

	// Bottom border
	buf.WriteString(tr.renderBottomBorder(colWidths))
	buf.WriteByte('\n')

	return buf.Bytes()
}

// calculateColumnWidths calculates optimal column widths
func (tr *TableRenderer) calculateColumnWidths() []int {
	if len(tr.rows) == 0 {
		return []int{}
	}

	// Reserve space for borders: | col1 | col2 | col3 |
	// That's: 1 + 1 + (numCols-1)*3 + 1 = numCols*3 + 1
	borderWidth := tr.maxColumns*3 + 1
	availableWidth := maxTableWidth - borderWidth

	// Divide available width equally among columns
	baseWidth := availableWidth / tr.maxColumns
	remainder := availableWidth % tr.maxColumns

	widths := make([]int, tr.maxColumns)
	for i := 0; i < tr.maxColumns; i++ {
		widths[i] = baseWidth
		if i < remainder {
			widths[i]++
		}
	}

	return widths
}

// renderRow renders a single table row
func (tr *TableRenderer) renderRow(row []string, colWidths []int, isHeader bool) string {
	var buf bytes.Buffer

	buf.WriteByte('|')
	for i := 0; i < tr.maxColumns; i++ {
		var cellText string
		if i < len(row) {
			cellText = row[i]
		} else {
			cellText = ""
		}

		// Truncate or pad cell text
		width := colWidths[i]
		if len(cellText) > width {
			cellText = cellText[:width]
		} else {
			// Apply alignment based on column alignment or default behavior
			var alignment extast.Alignment
			if i < len(tr.alignments) {
				alignment = tr.alignments[i]
			} else {
				alignment = extast.AlignNone
			}

			// Apply alignment
			if isHeader && alignment == extast.AlignNone {
				// Default: center-align headers if no specific alignment
				cellText = centerText(cellText, width)
			} else if alignment == extast.AlignRight {
				cellText = padLeft(cellText, width)
			} else if alignment == extast.AlignCenter {
				cellText = centerText(cellText, width)
			} else {
				// Left align (default for AlignLeft and AlignNone)
				cellText = padRight(cellText, width)
			}
		}

		buf.WriteByte(' ')
		buf.WriteString(cellText)
		buf.WriteString(" |")
	}

	return buf.String()
}

// renderTopBorder renders the top border
func (tr *TableRenderer) renderTopBorder(colWidths []int) string {
	var buf bytes.Buffer
	buf.WriteString("+")
	for _, width := range colWidths {
		buf.WriteString(strings.Repeat("-", width+2))
		buf.WriteString("+")
	}
	return buf.String()
}

// renderMiddleBorder renders the middle border (after header)
func (tr *TableRenderer) renderMiddleBorder(colWidths []int) string {
	var buf bytes.Buffer
	buf.WriteString("+")
	for _, width := range colWidths {
		buf.WriteString(strings.Repeat("-", width+2))
		buf.WriteString("+")
	}
	return buf.String()
}

// renderBottomBorder renders the bottom border
func (tr *TableRenderer) renderBottomBorder(colWidths []int) string {
	return tr.renderTopBorder(colWidths)
}

// renderReceiptFormat renders 2-column tables in simple receipt format (no borders)
func (tr *TableRenderer) renderReceiptFormat() []byte {
	var buf bytes.Buffer

	if len(tr.rows) == 0 {
		return buf.Bytes()
	}

	// Use full width for receipt format (no borders)
	col1Width := maxTableWidth - 12 // Reserve ~12 chars for prices
	col2Width := 10                 // Price column width

	for i, row := range tr.rows {
		if len(row) < 2 {
			continue
		}

		col1Text := row[0]
		col2Text := row[1]

		// Truncate if too long
		if len(col1Text) > col1Width {
			col1Text = col1Text[:col1Width]
		}
		if len(col2Text) > col2Width {
			col2Text = col2Text[:col2Width]
		}

		// Get alignment for each column
		var align1, align2 extast.Alignment
		if 0 < len(tr.alignments) {
			align1 = tr.alignments[0]
		}
		if 1 < len(tr.alignments) {
			align2 = tr.alignments[1]
		}

		// Apply alignment for column 1
		switch align1 {
		case extast.AlignRight:
			col1Text = padLeft(col1Text, col1Width)
		case extast.AlignCenter:
			col1Text = centerText(col1Text, col1Width)
		default:
			col1Text = padRight(col1Text, col1Width)
		}

		// Apply alignment for column 2 (typically right-aligned for prices)
		switch align2 {
		case extast.AlignLeft:
			col2Text = padRight(col2Text, col2Width)
		case extast.AlignCenter:
			col2Text = centerText(col2Text, col2Width)
		default:
			// Default to right-align for second column (prices)
			col2Text = padLeft(col2Text, col2Width)
		}

		// Add separator after header row
		if i == 0 {
			buf.WriteString(col1Text)
			buf.WriteString("  ")
			buf.WriteString(col2Text)
			buf.WriteByte('\n')
			// Use bitmap solid line (46 chars * 12 pixels = 552 pixels)
			buf.Write(escpos.SolidLine(552))
			buf.WriteByte('\n')
		} else {
			buf.WriteString(col1Text)
			buf.WriteString("  ")
			buf.WriteString(col2Text)
			buf.WriteByte('\n')
		}
	}

	return buf.Bytes()
}

// renderVerticalLayout renders tables with >3 columns in vertical key-value format
func (tr *TableRenderer) renderVerticalLayout() []byte {
	var buf bytes.Buffer

	if len(tr.rows) == 0 {
		return buf.Bytes()
	}

	// First row is treated as headers
	headers := tr.rows[0]

	// Render each data row as a vertical block
	for rowIdx := 1; rowIdx < len(tr.rows); rowIdx++ {
		row := tr.rows[rowIdx]

		buf.WriteString(fmt.Sprintf("--- Entry %d ---\n", rowIdx))

		for colIdx := 0; colIdx < len(headers); colIdx++ {
			header := ""
			value := ""

			if colIdx < len(headers) {
				header = headers[colIdx]
			}
			if colIdx < len(row) {
				value = row[colIdx]
			}

			buf.WriteString(fmt.Sprintf("%s: %s\n", header, value))
		}

		buf.WriteByte('\n')
	}

	return buf.Bytes()
}

// Helper functions

// extractCellText extracts text content from a table cell
func extractCellText(cell *extast.TableCell, source []byte) string {
	var buf bytes.Buffer

	// Walk through cell children to extract text
	for child := cell.FirstChild(); child != nil; child = child.NextSibling() {
		if text, ok := child.(*ast.Text); ok {
			buf.Write(text.Segment.Value(source))
		} else if str, ok := child.(*ast.String); ok {
			buf.Write(str.Value)
		} else {
			// Recursively extract text from nested elements
			extractTextRecursive(child, source, &buf)
		}
	}

	return strings.TrimSpace(buf.String())
}

// extractTextRecursive recursively extracts text from AST nodes
func extractTextRecursive(node ast.Node, source []byte, buf *bytes.Buffer) {
	if text, ok := node.(*ast.Text); ok {
		buf.Write(text.Segment.Value(source))
	} else if str, ok := node.(*ast.String); ok {
		buf.Write(str.Value)
	}

	// Recurse into children
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		extractTextRecursive(child, source, buf)
	}
}

// padRight pads text to the right with spaces (left-align)
func padRight(text string, width int) string {
	if len(text) >= width {
		return text
	}
	return text + strings.Repeat(" ", width-len(text))
}

// padLeft pads text to the left with spaces (right-align)
func padLeft(text string, width int) string {
	if len(text) >= width {
		return text
	}
	return strings.Repeat(" ", width-len(text)) + text
}

// centerText centers text within the given width
func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	leftPad := (width - len(text)) / 2
	rightPad := width - len(text) - leftPad
	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}
