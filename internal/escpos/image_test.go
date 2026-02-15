package escpos

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to create a test image file
func createTestImage(t *testing.T, width, height int, filename string) string {
	t.Helper()

	// Create a simple test image with a pattern
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a pattern: top half white, bottom half black
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if y < height/2 {
				img.Set(x, y, color.White)
			} else {
				img.Set(x, y, color.Black)
			}
		}
	}

	// Create temp directory
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, filename)

	// Save image
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("failed to close test image: %v", err)
		}
	}()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("failed to encode test image: %v", err)
	}

	return filePath
}

// Helper function to create a gradient test image
func createGradientImage(t *testing.T, width, height int, filename string) string {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create gradient from white to black
	for y := 0; y < height; y++ {
		grayValue := uint8(255 * y / height)
		for x := 0; x < width; x++ {
			img.Set(x, y, color.Gray{Y: 255 - grayValue})
		}
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create gradient image: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("failed to close gradient image: %v", err)
		}
	}()

	if err := png.Encode(file, img); err != nil {
		t.Fatalf("failed to encode gradient image: %v", err)
	}

	return filePath
}

func TestConvertImageToESCPOS(t *testing.T) {
	tests := []struct {
		name              string
		width             int
		height            int
		thresholdPercent  int
		expectedMinLength int
	}{
		{
			name:              "small image",
			width:             100,
			height:            100,
			thresholdPercent:  50,
			expectedMinLength: 100,
		},
		{
			name:              "medium image",
			width:             200,
			height:            150,
			thresholdPercent:  50,
			expectedMinLength: 200,
		},
		{
			name:              "wide image (should scale)",
			width:             500,
			height:            100,
			thresholdPercent:  50,
			expectedMinLength: 200,
		},
		{
			name:              "high threshold (more black)",
			width:             100,
			height:            100,
			thresholdPercent:  90,
			expectedMinLength: 100,
		},
		{
			name:              "low threshold (more white)",
			width:             100,
			height:            100,
			thresholdPercent:  10,
			expectedMinLength: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imagePath := createTestImage(t, tt.width, tt.height, "test.png")

			result, err := ConvertImageToESCPOS(imagePath, tt.thresholdPercent)
			if err != nil {
				t.Fatalf("ConvertImageToESCPOS() error = %v", err)
			}

			// Check that result is not empty
			if len(result) < tt.expectedMinLength {
				t.Errorf("ConvertImageToESCPOS() result too short, got %d bytes, want at least %d", len(result), tt.expectedMinLength)
			}

			// Check for center align at start
			if !bytes.HasPrefix(result, []byte{0x1B, 0x61, 0x01}) {
				t.Error("ConvertImageToESCPOS() should start with center align command")
			}

			// Check for bitmap print command (GS v 0)
			if !bytes.Contains(result, []byte{0x1D, 0x76, 0x30, 0x00}) {
				t.Error("ConvertImageToESCPOS() should contain bitmap print command")
			}

			// Check for reset align and line feed at end
			if !bytes.HasSuffix(result, []byte{0x1B, 0x61, 0x00, 0x0A}) {
				t.Error("ConvertImageToESCPOS() should end with align reset and line feed")
			}
		})
	}
}

func TestConvertImageToESCPOS_Scaling(t *testing.T) {
	// Test that images wider than 384 pixels are scaled down
	imagePath := createTestImage(t, 800, 600, "large.png")

	result, err := ConvertImageToESCPOS(imagePath, 50)
	if err != nil {
		t.Fatalf("ConvertImageToESCPOS() error = %v", err)
	}

	// Parse bitmap dimensions from ESC/POS command
	// Format: 0x1D 0x76 0x30 0x00 wL wH hL hH [bitmap data]
	if len(result) < 8 {
		t.Fatal("ConvertImageToESCPOS() result too short to contain bitmap header")
	}

	// Find bitmap command position
	bitmapCmdIndex := bytes.Index(result, []byte{0x1D, 0x76, 0x30, 0x00})
	if bitmapCmdIndex == -1 {
		t.Fatal("ConvertImageToESCPOS() missing bitmap command")
	}

	// Extract width in bytes (little-endian)
	widthBytesLow := result[bitmapCmdIndex+4]
	widthBytesHigh := result[bitmapCmdIndex+5]
	widthBytes := int(widthBytesLow) + int(widthBytesHigh)*256

	// Maximum width should be 384 pixels = 48 bytes
	if widthBytes > 48 {
		t.Errorf("ConvertImageToESCPOS() width = %d bytes, want <= 48 bytes (384 pixels)", widthBytes)
	}
}

func TestConvertImageToESCPOS_ThresholdEffect(t *testing.T) {
	// Test that threshold affects the bitmap output
	imagePath := createGradientImage(t, 100, 100, "gradient.png")

	lowThreshold, err := ConvertImageToESCPOS(imagePath, 10)
	if err != nil {
		t.Fatalf("ConvertImageToESCPOS(10) error = %v", err)
	}

	highThreshold, err := ConvertImageToESCPOS(imagePath, 90)
	if err != nil {
		t.Fatalf("ConvertImageToESCPOS(90) error = %v", err)
	}

	// The results should be different (different bitmap data)
	if bytes.Equal(lowThreshold, highThreshold) {
		t.Error("ConvertImageToESCPOS() threshold should affect output")
	}

	// High threshold should generally produce more data (more black pixels)
	// Note: This is not always guaranteed due to compression, but for a gradient it should hold
	if len(highThreshold) < len(lowThreshold) {
		t.Logf("Warning: High threshold produced less data than low threshold (may vary)")
	}
}

func TestConvertImageToESCPOS_InvalidFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
	}{
		{
			name:     "non-existent file",
			filePath: "/tmp/nonexistent_file_12345.png",
		},
		{
			name:     "empty path",
			filePath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ConvertImageToESCPOS(tt.filePath, 50)
			if err == nil {
				t.Error("ConvertImageToESCPOS() should return error for invalid file")
			}
		})
	}
}

func TestConvertImageToESCPOS_InvalidImageData(t *testing.T) {
	// Create a file with non-image data
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.png")

	err := os.WriteFile(invalidFile, []byte("this is not an image"), 0644)
	if err != nil {
		t.Fatalf("failed to create invalid file: %v", err)
	}

	_, err = ConvertImageToESCPOS(invalidFile, 50)
	if err == nil {
		t.Error("ConvertImageToESCPOS() should return error for invalid image data")
	}
}

func TestConvertImageToESCPOS_BitmapStructure(t *testing.T) {
	// Test the structure of the generated bitmap
	imagePath := createTestImage(t, 100, 100, "test.png")

	result, err := ConvertImageToESCPOS(imagePath, 50)
	if err != nil {
		t.Fatalf("ConvertImageToESCPOS() error = %v", err)
	}

	// Find bitmap command
	bitmapCmdIndex := bytes.Index(result, []byte{0x1D, 0x76, 0x30, 0x00})
	if bitmapCmdIndex == -1 {
		t.Fatal("ConvertImageToESCPOS() missing bitmap command")
	}

	// Extract dimensions
	widthBytesLow := result[bitmapCmdIndex+4]
	widthBytesHigh := result[bitmapCmdIndex+5]
	widthBytes := int(widthBytesLow) + int(widthBytesHigh)*256

	heightLow := result[bitmapCmdIndex+6]
	heightHigh := result[bitmapCmdIndex+7]
	height := int(heightLow) + int(heightHigh)*256

	// Calculate expected bitmap size
	expectedBitmapSize := widthBytes * height

	// Check that bitmap data exists
	bitmapStart := bitmapCmdIndex + 8
	bitmapEnd := bitmapStart + expectedBitmapSize

	if len(result) < bitmapEnd {
		t.Errorf("ConvertImageToESCPOS() bitmap data incomplete, got %d bytes, need %d", len(result), bitmapEnd)
	}
}

func TestConvertImageToESCPOS_EdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		width            int
		height           int
		thresholdPercent int
	}{
		{
			name:             "1x1 pixel",
			width:            1,
			height:           1,
			thresholdPercent: 50,
		},
		{
			name:             "very wide",
			width:            1000,
			height:           10,
			thresholdPercent: 50,
		},
		{
			name:             "very tall",
			width:            10,
			height:           1000,
			thresholdPercent: 50,
		},
		{
			name:             "threshold 0",
			width:            100,
			height:           100,
			thresholdPercent: 0,
		},
		{
			name:             "threshold 100",
			width:            100,
			height:           100,
			thresholdPercent: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imagePath := createTestImage(t, tt.width, tt.height, "test.png")

			result, err := ConvertImageToESCPOS(imagePath, tt.thresholdPercent)
			if err != nil {
				t.Fatalf("ConvertImageToESCPOS() error = %v", err)
			}

			if len(result) == 0 {
				t.Error("ConvertImageToESCPOS() returned empty result")
			}
		})
	}
}

// Benchmark tests
func BenchmarkConvertImageToESCPOS_Small(b *testing.B) {
	// Create a temporary test image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	tmpFile := filepath.Join(b.TempDir(), "bench.png")

	file, err := os.Create(tmpFile)
	if err != nil {
		b.Fatalf("failed to create benchmark file: %v", err)
	}
	if err := png.Encode(file, img); err != nil {
		b.Fatalf("failed to encode benchmark image: %v", err)
	}
	if err := file.Close(); err != nil {
		b.Fatalf("failed to close benchmark file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertImageToESCPOS(tmpFile, 50)
	}
}

func BenchmarkConvertImageToESCPOS_Large(b *testing.B) {
	// Create a temporary test image
	img := image.NewRGBA(image.Rect(0, 0, 500, 500))
	tmpFile := filepath.Join(b.TempDir(), "bench.png")

	file, err := os.Create(tmpFile)
	if err != nil {
		b.Fatalf("failed to create benchmark file: %v", err)
	}
	if err := png.Encode(file, img); err != nil {
		b.Fatalf("failed to encode benchmark image: %v", err)
	}
	if err := file.Close(); err != nil {
		b.Fatalf("failed to close benchmark file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertImageToESCPOS(tmpFile, 50)
	}
}

func BenchmarkConvertImageToESCPOS_Different_Thresholds(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	tmpFile := filepath.Join(b.TempDir(), "bench.png")

	file, err := os.Create(tmpFile)
	if err != nil {
		b.Fatalf("failed to create benchmark file: %v", err)
	}
	if err := png.Encode(file, img); err != nil {
		b.Fatalf("failed to encode benchmark image: %v", err)
	}
	if err := file.Close(); err != nil {
		b.Fatalf("failed to close benchmark file: %v", err)
	}

	thresholds := []int{0, 25, 50, 75, 100}

	for _, threshold := range thresholds {
		b.Run(string(rune(threshold)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ConvertImageToESCPOS(tmpFile, threshold)
			}
		})
	}
}
