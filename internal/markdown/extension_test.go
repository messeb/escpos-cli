package markdown

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func TestBarcodeParser_Trigger(t *testing.T) {
	bp := NewBarcodeParser()
	triggers := bp.Trigger()

	if len(triggers) != 1 {
		t.Errorf("Trigger() returned %d triggers, want 1", len(triggers))
	}

	if triggers[0] != '`' {
		t.Errorf("Trigger() returned %q, want %q", triggers[0], '`')
	}
}

func TestBarcodeParser_Open(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantBarcode  bool
		wantType     BarcodeType
		wantData     string
	}{
		{
			name: "QR code block",
			input: "```qr\nhttps://example.com\n```\n",
			wantBarcode: true,
			wantType: BarcodeQR,
			wantData: "https://example.com",
		},
		{
			name: "PDF417 barcode",
			input: "```pdf417\nTEST123\n```\n",
			wantBarcode: true,
			wantType: BarcodePDF417,
			wantData: "TEST123",
		},
		{
			name: "DataMatrix barcode",
			input: "```datamatrix\nABCDEF\n```\n",
			wantBarcode: true,
			wantType: BarcodeDataMatrix,
			wantData: "ABCDEF",
		},
		{
			name: "EAN13 barcode",
			input: "```ean13\n1234567890123\n```\n",
			wantBarcode: true,
			wantType: BarcodeEAN13,
			wantData: "1234567890123",
		},
		{
			name: "regular code block",
			input: "```javascript\nconsole.log('hello');\n```\n",
			wantBarcode: false,
		},
		{
			name: "not a code fence",
			input: "just some text\n",
			wantBarcode: false,
		},
		{
			name: "incomplete fence",
			input: "``qr\ntest\n```\n",
			wantBarcode: false,
		},
		{
			name: "multiline QR data",
			input: "```qr\nBEGIN:VCARD\nVERSION:3.0\nFN:John Doe\nEND:VCARD\n```\n",
			wantBarcode: true,
			wantType: BarcodeQR,
			wantData: "BEGIN:VCARD\nVERSION:3.0\nFN:John Doe\nEND:VCARD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := NewBarcodeParser()
			reader := text.NewReader([]byte(tt.input))
			pc := parser.NewContext()

			node, state := bp.Open(nil, reader, pc)

			if tt.wantBarcode {
				if node == nil {
					t.Error("Open() returned nil node, want BarcodeBlock")
					return
				}

				barcodeBlock, ok := node.(*BarcodeBlock)
				if !ok {
					t.Errorf("Open() returned %T, want *BarcodeBlock", node)
					return
				}

				if barcodeBlock.BarcodeType != tt.wantType {
					t.Errorf("BarcodeType = %v, want %v", barcodeBlock.BarcodeType, tt.wantType)
				}

				if barcodeBlock.Data != tt.wantData {
					t.Errorf("Data = %q, want %q", barcodeBlock.Data, tt.wantData)
				}

				if state != parser.NoChildren {
					t.Errorf("State = %v, want parser.NoChildren", state)
				}
			} else {
				if node != nil {
					t.Errorf("Open() returned %v, want nil", node)
				}
			}
		})
	}
}

func TestBarcodeParser_Continue(t *testing.T) {
	bp := NewBarcodeParser()
	reader := text.NewReader([]byte("test"))
	pc := parser.NewContext()

	state := bp.Continue(nil, reader, pc)
	if state != parser.Close {
		t.Errorf("Continue() = %v, want parser.Close", state)
	}
}

func TestBarcodeParser_CanInterruptParagraph(t *testing.T) {
	bp := NewBarcodeParser()
	if !bp.CanInterruptParagraph() {
		t.Error("CanInterruptParagraph() = false, want true")
	}
}

func TestBarcodeParser_CanAcceptIndentedLine(t *testing.T) {
	bp := NewBarcodeParser()
	if bp.CanAcceptIndentedLine() {
		t.Error("CanAcceptIndentedLine() = true, want false")
	}
}

func TestNewBarcodeBlock(t *testing.T) {
	tests := []struct {
		name        string
		barcodeType BarcodeType
		data        string
	}{
		{
			name:        "QR block",
			barcodeType: BarcodeQR,
			data:        "test data",
		},
		{
			name:        "PDF417 block",
			barcodeType: BarcodePDF417,
			data:        "12345",
		},
		{
			name:        "DataMatrix block",
			barcodeType: BarcodeDataMatrix,
			data:        "ABCDE",
		},
		{
			name:        "EAN13 block",
			barcodeType: BarcodeEAN13,
			data:        "1234567890123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := NewBarcodeBlock(tt.barcodeType, tt.data)

			if block.BarcodeType != tt.barcodeType {
				t.Errorf("BarcodeType = %v, want %v", block.BarcodeType, tt.barcodeType)
			}

			if block.Data != tt.data {
				t.Errorf("Data = %q, want %q", block.Data, tt.data)
			}

			// Test Kind() method
			kind := block.Kind()
			if kind.String() != "BarcodeBlock" {
				t.Errorf("Kind() = %v, want BarcodeBlock", kind)
			}
		})
	}
}

func TestBarcodeExtension_Extend(t *testing.T) {
	md := goldmark.New()
	ext := NewBarcodeExtension()

	// Extend should not panic
	ext.Extend(md)

	// Parse markdown with barcode
	source := []byte("```qr\ntest\n```\n")
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	// Walk the AST to find barcode block
	found := false
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if _, ok := n.(*BarcodeBlock); ok {
				found = true
			}
		}
		return ast.WalkContinue, nil
	})

	if !found {
		t.Error("BarcodeExtension did not parse barcode block")
	}
}

func TestBarcodeBlock_Dump(t *testing.T) {
	block := NewBarcodeBlock(BarcodeQR, "test")
	source := []byte("test source")

	// Dump should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Dump() panicked: %v", r)
		}
	}()

	block.Dump(source, 0)
}
