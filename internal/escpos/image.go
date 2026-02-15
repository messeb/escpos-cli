package escpos

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	_ "golang.org/x/image/bmp"
)

// ConvertImageToESCPOS converts an image file to ESC/POS bitmap format
// threshold is a percentage (0-100, higher = more black pixels)
func ConvertImageToESCPOS(imagePath string, thresholdPercent int) ([]byte, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	// Scale to max width of 384 pixels
	maxWidth := 384
	if width > maxWidth {
		ratio := float64(maxWidth) / float64(width)
		width = maxWidth
		height = int(float64(height) * ratio)
	}

	widthBytes := (width + 7) / 8
	bitmap := make([]byte, widthBytes*height)

	// Convert threshold percentage to 16-bit grayscale value
	threshold := uint32(thresholdPercent * 655)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := x * bounds.Max.X / width
			srcY := y * bounds.Max.Y / height

			r, g, b, _ := img.At(srcX, srcY).RGBA()
			gray := (r + g + b) / 3

			if gray < threshold {
				byteIndex := y*widthBytes + x/8
				bitIndex := 7 - (x % 8)
				bitmap[byteIndex] |= 1 << bitIndex
			}
		}
	}

	var buf []byte
	buf = append(buf, 0x1B, 0x61, 0x01)       // Center align
	buf = append(buf, 0x1D, 0x76, 0x30, 0x00) // GS v 0 - Print bitmap
	buf = append(buf, byte(widthBytes%256), byte(widthBytes/256))
	buf = append(buf, byte(height%256), byte(height/256))
	buf = append(buf, bitmap...)
	buf = append(buf, 0x1B, 0x61, 0x00, 0x0A) // Reset align, line feed

	return buf, nil
}
