package escpos

// ESC/POS command helper functions
// These functions return byte slices for ESC/POS protocol commands

// Init initializes the printer
func Init() []byte {
	return []byte{0x1B, 0x40}
}

// LineFeed sends a line feed with carriage return (newline + return to column 0)
func LineFeed() []byte {
	return []byte{0x0D, 0x0A} // CR+LF for proper line break
}

// Feed feeds n lines
func Feed(n int) []byte {
	return []byte{0x1B, 0x64, byte(n)}
}

// Cut cuts the paper
func Cut() []byte {
	return []byte{0x1D, 0x56, 0x01}
}

// AlignCenter sets center alignment
func AlignCenter() []byte {
	return []byte{0x1B, 0x61, 0x01}
}

// AlignLeft sets left alignment
func AlignLeft() []byte {
	return []byte{0x1B, 0x61, 0x00}
}

// BoldOn enables bold text
func BoldOn() []byte {
	return []byte{0x1B, 0x45, 0x01}
}

// BoldOff disables bold text
func BoldOff() []byte {
	return []byte{0x1B, 0x45, 0x00}
}

// UnderlineOn enables underline
func UnderlineOn() []byte {
	return []byte{0x1B, 0x2D, 0x01}
}

// UnderlineOff disables underline
func UnderlineOff() []byte {
	return []byte{0x1B, 0x2D, 0x00}
}

// DoubleSize sets double width and height
func DoubleSize() []byte {
	return []byte{0x1B, 0x21, 0x30}
}

// DoubleHeight sets double height only
func DoubleHeight() []byte {
	return []byte{0x1B, 0x21, 0x10}
}

// DoubleWidth sets double width only
func DoubleWidth() []byte {
	return []byte{0x1B, 0x21, 0x20}
}

// NormalSize resets to normal text size
func NormalSize() []byte {
	return []byte{0x1B, 0x21, 0x00}
}

// SolidLine generates a solid line bitmap
func SolidLine(widthPixels int) []byte {
	widthBytes := (widthPixels + 7) / 8
	bitmap := make([]byte, widthBytes)
	for i := range bitmap {
		bitmap[i] = 0xFF // All bits set (solid line)
	}

	var buf []byte
	buf = append(buf, 0x1D, 0x76, 0x30, 0x00) // GS v 0
	buf = append(buf, byte(widthBytes%256), byte(widthBytes/256))
	buf = append(buf, byte(1), byte(0)) // Height = 1 pixel
	buf = append(buf, bitmap...)
	return buf
}
