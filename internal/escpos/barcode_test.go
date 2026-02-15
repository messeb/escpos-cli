package escpos

import (
	"bytes"
	"testing"
)

func TestGenerateQRCode(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		shouldPanic bool
	}{
		{
			name:        "simple text",
			data:        "Hello World",
			shouldPanic: false,
		},
		{
			name:        "URL",
			data:        "https://example.com",
			shouldPanic: false,
		},
		{
			name:        "numeric data",
			data:        "123456789",
			shouldPanic: false,
		},
		{
			name:        "long text",
			data:        "This is a very long text that should still generate a valid QR code with proper encoding",
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateQRCode(tt.data)

			// Check that result is not empty
			if len(result) == 0 {
				t.Error("GenerateQRCode() returned empty result")
			}

			// Check for center align command at start
			if !bytes.HasPrefix(result, []byte{0x1B, 0x61, 0x01}) {
				t.Error("GenerateQRCode() should start with center align command")
			}

			// Check for bitmap print command (GS v 0)
			if !bytes.Contains(result, []byte{0x1D, 0x76, 0x30, 0x00}) {
				t.Error("GenerateQRCode() should contain bitmap print command")
			}

			// Check for reset align at end
			if !bytes.Contains(result, []byte{0x1B, 0x61, 0x00}) {
				t.Error("GenerateQRCode() should contain align reset command")
			}

			// Check for line feed at end
			if result[len(result)-1] != 0x0A {
				t.Error("GenerateQRCode() should end with line feed")
			}
		})
	}
}

func TestGeneratePDF417(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "simple text",
			data: "TEST123",
		},
		{
			name: "URL",
			data: "https://example.com/order/12345",
		},
		{
			name: "alphanumeric",
			data: "ABC123XYZ",
		},
		{
			name: "single character",
			data: "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GeneratePDF417(tt.data)

			// Check that result is not empty
			if len(result) == 0 {
				t.Error("GeneratePDF417() returned empty result")
			}

			// Check for center align at start
			if !bytes.HasPrefix(result, []byte{0x1B, 0x61, 0x01}) {
				t.Error("GeneratePDF417() should start with center align command")
			}

			// Check for GS ( k commands (0x1D, 0x28, 0x6B)
			if !bytes.Contains(result, []byte{0x1D, 0x28, 0x6B}) {
				t.Error("GeneratePDF417() should contain GS ( k commands")
			}

			// Check for data in the result
			if !bytes.Contains(result, []byte(tt.data)) {
				t.Errorf("GeneratePDF417() should contain data: %s", tt.data)
			}

			// Check for reset align at end
			if !bytes.HasSuffix(result, []byte{0x1B, 0x61, 0x00}) {
				t.Error("GeneratePDF417() should end with align reset command")
			}

			// Check for line feeds before reset
			if !bytes.Contains(result, []byte{0x0A, 0x0A}) {
				t.Error("GeneratePDF417() should contain double line feed")
			}
		})
	}
}

func TestGeneratePDF417Commands(t *testing.T) {
	// Test specific command structure
	result := GeneratePDF417("TEST")

	// Expected command sequence components
	expectedCommands := []struct {
		name    string
		pattern []byte
	}{
		{
			name:    "set columns",
			pattern: []byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 48, 65}, // GS ( k pL pH cn fn
		},
		{
			name:    "set rows",
			pattern: []byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 48, 66},
		},
		{
			name:    "set width",
			pattern: []byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 48, 67},
		},
		{
			name:    "set row height",
			pattern: []byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 48, 68},
		},
		{
			name:    "set error correction",
			pattern: []byte{0x1D, 0x28, 0x6B, 0x04, 0x00, 48, 69},
		},
	}

	for _, cmd := range expectedCommands {
		t.Run(cmd.name, func(t *testing.T) {
			if !bytes.Contains(result, cmd.pattern) {
				t.Errorf("GeneratePDF417() missing %s command: %v", cmd.name, cmd.pattern)
			}
		})
	}
}

func TestGenerateDataMatrix(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{
			name: "simple text",
			data: "TEST",
		},
		{
			name: "URL",
			data: "https://example.com",
		},
		{
			name: "numeric",
			data: "987654321",
		},
		{
			name: "alphanumeric",
			data: "ABC123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateDataMatrix(tt.data)

			// Check that result is not empty
			if len(result) == 0 {
				t.Error("GenerateDataMatrix() returned empty result")
			}

			// Check for center align at start
			if !bytes.HasPrefix(result, []byte{0x1B, 0x61, 0x01}) {
				t.Error("GenerateDataMatrix() should start with center align command")
			}

			// Check for GS ( k commands
			if !bytes.Contains(result, []byte{0x1D, 0x28, 0x6B}) {
				t.Error("GenerateDataMatrix() should contain GS ( k commands")
			}

			// Check for DataMatrix symbol type (54)
			if !bytes.Contains(result, []byte{54}) {
				t.Error("GenerateDataMatrix() should contain DataMatrix symbol type")
			}

			// Check for data in the result
			if !bytes.Contains(result, []byte(tt.data)) {
				t.Errorf("GenerateDataMatrix() should contain data: %s", tt.data)
			}

			// Check for reset align at end
			if !bytes.HasSuffix(result, []byte{0x1B, 0x61, 0x00}) {
				t.Error("GenerateDataMatrix() should end with align reset command")
			}
		})
	}
}

func TestGenerateDataMatrixCommands(t *testing.T) {
	// Test specific command structure
	result := GenerateDataMatrix("TEST")

	// Check module size command
	moduleSizeCmd := []byte{0x1D, 0x28, 0x6B, 3, 0, 54, 50, 6}
	if !bytes.Contains(result, moduleSizeCmd) {
		t.Error("GenerateDataMatrix() missing module size command")
	}

	// Check print command
	printCmd := []byte{0x1D, 0x28, 0x6B, 3, 0, 54, 81, 48}
	if !bytes.Contains(result, printCmd) {
		t.Error("GenerateDataMatrix() missing print command")
	}
}

func TestGenerateEAN13(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		shouldRun bool
	}{
		{
			name:      "valid 12 digits",
			code:      "123456789012",
			shouldRun: true,
		},
		{
			name:      "valid 13 digits",
			code:      "1234567890123",
			shouldRun: true,
		},
		{
			name:      "short code",
			code:      "12345",
			shouldRun: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateEAN13(tt.code)

			// Check that result is not empty
			if len(result) == 0 {
				t.Error("GenerateEAN13() returned empty result")
			}

			// Check for center align at start
			if !bytes.HasPrefix(result, []byte{0x1B, 0x61, 0x01}) {
				t.Error("GenerateEAN13() should start with center align command")
			}

			// Check for barcode height command (GS h)
			if !bytes.Contains(result, []byte{0x1D, 0x68}) {
				t.Error("GenerateEAN13() should contain height command")
			}

			// Check for barcode width command (GS w)
			if !bytes.Contains(result, []byte{0x1D, 0x77}) {
				t.Error("GenerateEAN13() should contain width command")
			}

			// Check for HRI position command (GS H)
			if !bytes.Contains(result, []byte{0x1D, 0x48}) {
				t.Error("GenerateEAN13() should contain HRI position command")
			}

			// Check for HRI font command (GS f)
			if !bytes.Contains(result, []byte{0x1D, 0x66}) {
				t.Error("GenerateEAN13() should contain HRI font command")
			}

			// Check for barcode print command (GS k)
			if !bytes.Contains(result, []byte{0x1D, 0x6B, 0x43}) {
				t.Error("GenerateEAN13() should contain barcode print command (EAN13)")
			}

			// Check for data in the result
			if !bytes.Contains(result, []byte(tt.code)) {
				t.Errorf("GenerateEAN13() should contain code: %s", tt.code)
			}

			// Check for reset align at end
			if !bytes.HasSuffix(result, []byte{0x1B, 0x61, 0x00}) {
				t.Error("GenerateEAN13() should end with align reset command")
			}
		})
	}
}

func TestGenerateEAN13CommandStructure(t *testing.T) {
	// Test specific EAN13 command parameters
	result := GenerateEAN13("123456789012")

	// Check barcode height (should be 0x64 = 100)
	heightCmd := []byte{0x1D, 0x68, 0x64}
	if !bytes.Contains(result, heightCmd) {
		t.Error("GenerateEAN13() should set height to 100")
	}

	// Check barcode width (should be 0x02)
	widthCmd := []byte{0x1D, 0x77, 0x02}
	if !bytes.Contains(result, widthCmd) {
		t.Error("GenerateEAN13() should set width to 2")
	}

	// Check HRI position (should be 0x02 = below)
	hriPosCmd := []byte{0x1D, 0x48, 0x02}
	if !bytes.Contains(result, hriPosCmd) {
		t.Error("GenerateEAN13() should set HRI position to below")
	}

	// Check HRI font (should be 0x00 = Font A)
	hriFontCmd := []byte{0x1D, 0x66, 0x00}
	if !bytes.Contains(result, hriFontCmd) {
		t.Error("GenerateEAN13() should set HRI font to A")
	}
}

// Benchmark tests
func BenchmarkGenerateQRCode(b *testing.B) {
	data := "https://example.com/order/12345"
	for i := 0; i < b.N; i++ {
		GenerateQRCode(data)
	}
}

func BenchmarkGeneratePDF417(b *testing.B) {
	data := "TEST123456"
	for i := 0; i < b.N; i++ {
		GeneratePDF417(data)
	}
}

func BenchmarkGenerateDataMatrix(b *testing.B) {
	data := "TEST123456"
	for i := 0; i < b.N; i++ {
		GenerateDataMatrix(data)
	}
}

func BenchmarkGenerateEAN13(b *testing.B) {
	code := "123456789012"
	for i := 0; i < b.N; i++ {
		GenerateEAN13(code)
	}
}

// Test barcode generation with different data sizes
func TestGenerateQRCodeDataSizes(t *testing.T) {
	sizes := []int{10, 50, 100, 500, 1000}

	for _, size := range sizes {
		t.Run(string(rune(size)), func(t *testing.T) {
			// Create data of specified size
			data := make([]byte, size)
			for i := range data {
				data[i] = 'A' + byte(i%26)
			}

			result := GenerateQRCode(string(data))

			if len(result) == 0 {
				t.Errorf("GenerateQRCode() failed with data size %d", size)
			}
		})
	}
}
