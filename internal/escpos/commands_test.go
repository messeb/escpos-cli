package escpos

import (
	"bytes"
	"testing"
)

func TestInit(t *testing.T) {
	got := Init()
	want := []byte{0x1B, 0x40}

	if !bytes.Equal(got, want) {
		t.Errorf("Init() = %v, want %v", got, want)
	}
}

func TestLineFeed(t *testing.T) {
	got := LineFeed()
	want := []byte{0x0D, 0x0A}

	if !bytes.Equal(got, want) {
		t.Errorf("LineFeed() = %v, want %v", got, want)
	}
}

func TestFeed(t *testing.T) {
	tests := []struct {
		name  string
		lines int
		want  []byte
	}{
		{
			name:  "feed 1 line",
			lines: 1,
			want:  []byte{0x1B, 0x64, 0x01},
		},
		{
			name:  "feed 5 lines",
			lines: 5,
			want:  []byte{0x1B, 0x64, 0x05},
		},
		{
			name:  "feed 10 lines",
			lines: 10,
			want:  []byte{0x1B, 0x64, 0x0A},
		},
		{
			name:  "feed 0 lines",
			lines: 0,
			want:  []byte{0x1B, 0x64, 0x00},
		},
		{
			name:  "feed 255 lines (max byte value)",
			lines: 255,
			want:  []byte{0x1B, 0x64, 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Feed(tt.lines)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("Feed(%d) = %v, want %v", tt.lines, got, tt.want)
			}
		})
	}
}

func TestCut(t *testing.T) {
	got := Cut()
	want := []byte{0x1D, 0x56, 0x01}

	if !bytes.Equal(got, want) {
		t.Errorf("Cut() = %v, want %v", got, want)
	}
}

func TestAlignment(t *testing.T) {
	tests := []struct {
		name string
		fn   func() []byte
		want []byte
	}{
		{
			name: "align center",
			fn:   AlignCenter,
			want: []byte{0x1B, 0x61, 0x01},
		},
		{
			name: "align left",
			fn:   AlignLeft,
			want: []byte{0x1B, 0x61, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestBold(t *testing.T) {
	tests := []struct {
		name string
		fn   func() []byte
		want []byte
	}{
		{
			name: "bold on",
			fn:   BoldOn,
			want: []byte{0x1B, 0x45, 0x01},
		},
		{
			name: "bold off",
			fn:   BoldOff,
			want: []byte{0x1B, 0x45, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestUnderline(t *testing.T) {
	tests := []struct {
		name string
		fn   func() []byte
		want []byte
	}{
		{
			name: "underline on",
			fn:   UnderlineOn,
			want: []byte{0x1B, 0x2D, 0x01},
		},
		{
			name: "underline off",
			fn:   UnderlineOff,
			want: []byte{0x1B, 0x2D, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestTextSize(t *testing.T) {
	tests := []struct {
		name string
		fn   func() []byte
		want []byte
	}{
		{
			name: "double size",
			fn:   DoubleSize,
			want: []byte{0x1B, 0x21, 0x30},
		},
		{
			name: "double height",
			fn:   DoubleHeight,
			want: []byte{0x1B, 0x21, 0x10},
		},
		{
			name: "double width",
			fn:   DoubleWidth,
			want: []byte{0x1B, 0x21, 0x20},
		},
		{
			name: "normal size",
			fn:   NormalSize,
			want: []byte{0x1B, 0x21, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if !bytes.Equal(got, tt.want) {
				t.Errorf("%s() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestSolidLine(t *testing.T) {
	tests := []struct {
		name        string
		widthPixels int
		wantHeader  []byte
		wantWidth   int // width in bytes
	}{
		{
			name:        "8 pixels (1 byte)",
			widthPixels: 8,
			wantHeader:  []byte{0x1D, 0x76, 0x30, 0x00},
			wantWidth:   1,
		},
		{
			name:        "16 pixels (2 bytes)",
			widthPixels: 16,
			wantHeader:  []byte{0x1D, 0x76, 0x30, 0x00},
			wantWidth:   2,
		},
		{
			name:        "384 pixels (48 bytes)",
			widthPixels: 384,
			wantHeader:  []byte{0x1D, 0x76, 0x30, 0x00},
			wantWidth:   48,
		},
		{
			name:        "9 pixels (rounds up to 2 bytes)",
			widthPixels: 9,
			wantHeader:  []byte{0x1D, 0x76, 0x30, 0x00},
			wantWidth:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SolidLine(tt.widthPixels)

			// Check header (GS v 0)
			if !bytes.HasPrefix(got, tt.wantHeader) {
				t.Errorf("SolidLine(%d) header = %v, want %v", tt.widthPixels, got[:4], tt.wantHeader)
			}

			// Check width bytes encoding (little-endian)
			widthLow := got[4]
			widthHigh := got[5]
			actualWidth := int(widthLow) + int(widthHigh)*256

			if actualWidth != tt.wantWidth {
				t.Errorf("SolidLine(%d) width = %d bytes, want %d bytes", tt.widthPixels, actualWidth, tt.wantWidth)
			}

			// Check height (should be 1)
			heightLow := got[6]
			heightHigh := got[7]
			height := int(heightLow) + int(heightHigh)*256

			if height != 1 {
				t.Errorf("SolidLine(%d) height = %d, want 1", tt.widthPixels, height)
			}

			// Check bitmap data (all 0xFF for solid line)
			bitmapStart := 8
			bitmapEnd := bitmapStart + tt.wantWidth
			for i := bitmapStart; i < bitmapEnd; i++ {
				if got[i] != 0xFF {
					t.Errorf("SolidLine(%d) bitmap[%d] = 0x%02X, want 0xFF", tt.widthPixels, i-bitmapStart, got[i])
				}
			}

			// Check total length
			expectedLen := 8 + tt.wantWidth
			if len(got) != expectedLen {
				t.Errorf("SolidLine(%d) length = %d, want %d", tt.widthPixels, len(got), expectedLen)
			}
		})
	}
}

// Benchmark tests for performance measurement
func BenchmarkInit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Init()
	}
}

func BenchmarkSolidLine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SolidLine(384)
	}
}

func BenchmarkTextFormatting(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BoldOn()
		DoubleSize()
		BoldOff()
		NormalSize()
	}
}
