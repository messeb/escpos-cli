package escpos

import (
	"github.com/skip2/go-qrcode"
)

// GenerateQRCode generates a QR code and converts it to ESC/POS bitmap format
func GenerateQRCode(data string) []byte {
	qr, _ := qrcode.New(data, qrcode.Medium)
	qr.DisableBorder = false
	img := qr.Image(200)

	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y
	widthBytes := (width + 7) / 8
	bitmap := make([]byte, widthBytes*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := (r + g + b) / 3
			if gray < 32768 {
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

	return buf
}

// GeneratePDF417 generates a PDF417 2D barcode using ESC/POS commands
func GeneratePDF417(data string) []byte {
	var buf []byte

	// Center align
	buf = append(buf, 0x1B, 0x61, 0x01)

	cn := byte(48) // '0'

	// Set PDF417 columns (1-30)
	pL := byte(3)
	pH := byte(0)
	fn := byte(65) // 'A' - Set columns
	n := byte(0)   // Auto
	buf = append(buf, 0x1D, 0x28, 0x6B, pL, pH, cn, fn, n)

	// Set PDF417 rows (3-90)
	pL = byte(3)
	pH = byte(0)
	fn = byte(66) // 'B' - Set rows
	n = byte(0)   // Auto
	buf = append(buf, 0x1D, 0x28, 0x6B, pL, pH, cn, fn, n)

	// Set PDF417 width (2-8)
	pL = byte(3)
	pH = byte(0)
	fn = byte(67) // 'C' - Set width
	n = byte(3)
	buf = append(buf, 0x1D, 0x28, 0x6B, pL, pH, cn, fn, n)

	// Set PDF417 row height (2-8)
	pL = byte(3)
	pH = byte(0)
	fn = byte(68) // 'D' - Set row height
	n = byte(3)
	buf = append(buf, 0x1D, 0x28, 0x6B, pL, pH, cn, fn, n)

	// Set error correction level (0-8)
	pL = byte(4)
	pH = byte(0)
	fn = byte(69) // 'E' - Set error correction
	n = byte(1)
	m := byte(1) // Level 1
	buf = append(buf, 0x1D, 0x28, 0x6B, pL, pH, cn, fn, n, m)

	// Store data
	dataLen := len(data)
	pL = byte((dataLen + 3) % 256)
	pH = byte((dataLen + 3) / 256)
	fn = byte(80) // 'P' - Store data
	m = byte(48)  // '0'
	buf = append(buf, 0x1D, 0x28, 0x6B, pL, pH, cn, fn, m)
	buf = append(buf, []byte(data)...)

	// Print PDF417
	pL = byte(3)
	pH = byte(0)
	fn = byte(81) // 'Q' - Print
	m = byte(48)  // '0'
	buf = append(buf, 0x1D, 0x28, 0x6B, pL, pH, cn, fn, m)

	buf = append(buf, 0x0A, 0x0A)

	// Reset align
	buf = append(buf, 0x1B, 0x61, 0x00)

	return buf
}

// GenerateDataMatrix generates a DataMatrix 2D barcode using ESC/POS commands
func GenerateDataMatrix(data string) []byte {
	var buf []byte

	// Center align
	buf = append(buf, 0x1B, 0x61, 0x01)

	cn := byte(54) // DataMatrix symbol type (0x36)

	// Set module size (fn 50)
	buf = append(buf, 0x1D, 0x28, 0x6B, 3, 0, cn, 50, 6) // module size 6

	// Store data: GS ( k pL pH cn 80 48 data
	dataLen := len(data) + 3
	buf = append(buf, 0x1D, 0x28, 0x6B, byte(dataLen%256), byte(dataLen/256), cn, 80, 48)
	buf = append(buf, []byte(data)...)

	// Print symbol: GS ( k 3 0 cn 81 48
	buf = append(buf, 0x1D, 0x28, 0x6B, 3, 0, cn, 81, 48)

	buf = append(buf, 0x0A, 0x0A)

	// Reset align
	buf = append(buf, 0x1B, 0x61, 0x00)

	return buf
}

// GenerateEAN13 generates an EAN-13 barcode using native ESC/POS barcode commands
// The data should be exactly 12 or 13 digits. If 12 digits, the check digit is calculated automatically by the printer.
func GenerateEAN13(code string) []byte {
	var buf []byte

	// Center align
	buf = append(buf, 0x1B, 0x61, 0x01)

	// Set barcode height (default 162)
	buf = append(buf, 0x1D, 0x68, 0x64) // Height = 100 dots

	// Set barcode width
	buf = append(buf, 0x1D, 0x77, 0x02) // Width = 2 (1-6)

	// HRI position (Human Readable Interpretation)
	buf = append(buf, 0x1D, 0x48, 0x02) // 0=none, 1=above, 2=below, 3=both

	// HRI font
	buf = append(buf, 0x1D, 0x66, 0x00) // Font A

	// Print barcode - EAN13
	buf = append(buf, 0x1D, 0x6B, 0x43)  // GS k 67 (EAN13)
	buf = append(buf, byte(len(code)))   // Length
	buf = append(buf, []byte(code)...)   // Data

	buf = append(buf, 0x0A, 0x0A)

	// Reset align
	buf = append(buf, 0x1B, 0x61, 0x00)

	return buf
}
