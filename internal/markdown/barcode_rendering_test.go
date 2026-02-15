package markdown

import (
	"bytes"
	"testing"

	"github.com/messeb/escpos-cli/internal/escpos"
	"github.com/yuin/goldmark/ast"
)

func TestESCPOSRenderer_renderBarcodeBlock(t *testing.T) {
	tests := []struct {
		name        string
		barcodeType BarcodeType
		data        string
		wantContain bool
	}{
		{
			name:        "QR code",
			barcodeType: BarcodeQR,
			data:        "https://example.com",
			wantContain: true,
		},
		{
			name:        "PDF417",
			barcodeType: BarcodePDF417,
			data:        "TEST123",
			wantContain: true,
		},
		{
			name:        "DataMatrix",
			barcodeType: BarcodeDataMatrix,
			data:        "ABC123",
			wantContain: true,
		},
		{
			name:        "EAN13",
			barcodeType: BarcodeEAN13,
			data:        "1234567890123",
			wantContain: true,
		},
		{
			name:        "unknown barcode type",
			barcodeType: BarcodeType(999),
			data:        "test",
			wantContain: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewESCPOSRenderer("/tmp", 75)
			r.currentColumn = 0

			block := NewBarcodeBlock(tt.barcodeType, tt.data)

			status, err := r.renderBarcodeBlock(block, true)
			if err != nil {
				t.Errorf("renderBarcodeBlock() error = %v", err)
			}

			if status != ast.WalkSkipChildren {
				t.Errorf("renderBarcodeBlock() status = %v, want WalkSkipChildren", status)
			}

			output := r.buf.Bytes()

			if tt.wantContain {
				// Should have generated some output
				if len(output) == 0 {
					t.Error("renderBarcodeBlock() produced no output")
				}

				// Should end with line feed
				if !bytes.HasSuffix(output, escpos.LineFeed()) {
					t.Error("renderBarcodeBlock() should end with line feed")
				}

				// Column should be reset
				if r.currentColumn != 0 {
					t.Errorf("renderBarcodeBlock() currentColumn = %d, want 0", r.currentColumn)
				}
			} else {
				// Unknown type should produce error message
				if !bytes.Contains(output, []byte("Unknown barcode type")) {
					t.Error("renderBarcodeBlock() should produce error message for unknown type")
				}
			}
		})
	}
}

func TestESCPOSRenderer_renderBarcodeBlock_exiting(t *testing.T) {
	r := NewESCPOSRenderer("/tmp", 75)
	block := NewBarcodeBlock(BarcodeQR, "test")

	status, err := r.renderBarcodeBlock(block, false)
	if err != nil {
		t.Errorf("renderBarcodeBlock(exiting) error = %v", err)
	}

	if status != ast.WalkSkipChildren {
		t.Errorf("renderBarcodeBlock(exiting) status = %v, want WalkSkipChildren", status)
	}

	// Should not produce output when exiting
	if r.buf.Len() > 0 {
		t.Error("renderBarcodeBlock(exiting) should not produce output")
	}
}
