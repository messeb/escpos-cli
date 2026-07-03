# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go application for printing markdown to a POS-80 thermal printer using the ESC/POS protocol. The printer is accessed via macOS CUPS (lpr command) rather than direct serial communication.

## Build and Run

**Build:**
```bash
go build -o escpos cmd/escpos/main.go
```

**Run commands:**
```bash
./escpos -file document.md                           # Use default printer (Printer_POS_80)
./escpos -printer MyPrinter -file document.md        # Specify printer
./escpos -file document.md -output receipt.bin       # Write raw ESC/POS to file (bypasses CUPS)
./escpos -file document.md -threshold 80             # Adjust image darkness
./escpos document.md                                 # Positional argument
```

**Output modes:**
- Default: renders and sends to the printer via CUPS (`lpr`).
- `-output <file>`: writes the raw ESC/POS byte stream to `<file>` and skips CUPS
  entirely. Useful on platforms without CUPS (e.g. Windows) or to capture bytes
  for a custom transport. `-printer` is ignored in this mode.
## Architecture

### Project Structure

The application uses a standard Go project layout with `cmd/` and `internal/` packages:

```
printer/
├── cmd/
│   └── escpos/            # CLI entry point
│       └── main.go        # Flag parsing, calls markdown.BuildMarkdownPrint()
├── internal/
│   ├── escpos/            # ESC/POS protocol primitives
│   │   ├── commands.go    # Command constants (Init, Bold, Cut, etc.)
│   │   ├── printer.go     # Print job management (SendToPrinter)
│   │   ├── image.go       # ConvertImageToESCPOS()
│   │   └── barcode.go     # GenerateQRCode(), GeneratePDF417(), GenerateDataMatrix()
│   └── markdown/          # Markdown rendering engine
│       ├── builder.go     # BuildMarkdownPrint() - entry point
│       ├── renderer.go    # ESCPOSRenderer - AST walker
│       ├── table.go       # TableRenderer - table layout
│       └── extension.go   # BarcodeExtension for Goldmark
├── testdata/              # Test markdown files
└── go.mod
```

### Package Responsibilities

**cmd/escpos/main.go:**
- CLI flag parsing (`-printer`, `-file`, `-threshold`)
- Calls `markdown.BuildMarkdownPrint()`
- Calls `escpos.SendToPrinter()`
- Minimal logic - just coordination

**internal/escpos:**
- ESC/POS protocol primitives and command constants
- Print job management (cancel queue, write temp file, lpr, cleanup)
- Image processing and barcode generation
- No rendering logic - just primitives used by markdown package

**internal/markdown:**
- Markdown parsing and rendering to ESC/POS byte sequences
- Goldmark integration with GFM and barcode extensions
- AST walking and node rendering
- Table layout algorithms
- Uses internal/escpos for low-level primitives

### ESC/POS Protocol Implementation
The application builds raw byte arrays using ESC/POS control codes:
- `0x1B, 0x40` - Initialize printer
- `0x1B, 0x61, 0x01` - Center align (0x00 = left, 0x02 = right)
- `0x1B, 0x45, 0x01` - Bold on (0x00 = off)
- `0x1B, 0x21, 0x30` - Double size text
- `0x1D, 0x76, 0x30, 0x00` - Print bitmap (GS v 0)
- `0x1D, 0x6B` - Print barcode (GS k)
- `0x1D, 0x56, 0x01` - Cut paper
- `0x1B, 0x64, 0x05` - Feed 5 lines

All text and formatting is done by appending bytes to a buffer, not using high-level APIs.


### Printing Workflow
1. Cancel any stuck print jobs: `cancel -a`
2. Build ESC/POS byte array from markdown
3. Write to temporary file `/tmp/print_<timestamp>.bin`
4. Send to printer via lpr: `lpr -P <printer-name> -o raw <file>`
5. Clean up temp file and cancel queue again

The printer name defaults to "Printer_POS_80" but can be configured via the `-printer` flag. The printer must be configured in macOS CUPS.

### Barcode/QR Code Generation
- **QR codes**: Generated using `github.com/skip2/go-qrcode` library, converted to bitmap, embedded in ESC/POS
- **EAN-13 barcodes**: Uses native ESC/POS barcode commands (GS k 67) - requires 12-13 numeric digits
- **PDF417 barcodes**: Uses ESC/POS 2D barcode commands (GS ( k) - printer support varies
- **DataMatrix**: Uses ESC/POS 2D barcode commands (GS ( k) - printer support varies

### Markdown Printing
The application supports printing markdown files with automatic conversion to ESC/POS format. This allows users to create print templates without writing code.

**Architecture**:
1. Parse markdown file using Goldmark parser (with GFM and custom barcode extensions)
2. Walk AST and convert each node to ESC/POS commands
3. Handle thermal printer constraints (40-char width, monospace font)
4. Generate final ESC/POS byte array

**Supported Markdown Features**:

*Text Formatting:*
- Headings (H1, H2, H3) → Double-size, double-height, double-width
- Bold (`**text**`) → ESC/POS bold command
- Italic (`*text*`) → Underline (thermal printers don't support italic)
- Paragraphs → Normal text with line breaks

*Lists:*
- Unordered lists → Bullets (•, ◦, ▪) based on nesting level
- Ordered lists → Numbered (1., 2., 3.)
- Task lists (`- [ ]` / `- [x]`) → Checkbox characters
- Nested lists supported up to 3 levels with indentation

*Images:*
- `![alt](path.png)` → Converts image to 1-bit bitmap, scales to 384px max width
- Relative paths resolved from markdown file location
- Supports PNG, JPG, BMP formats
- Threshold parameter controls darkness (0-100, higher = more black pixels)

*Barcodes:*
- QR codes: ` ```qr \n data \n ``` ` → Generates QR code bitmap
- EAN-13: ` ```ean13 \n 1234567890128 \n ``` ` → Uses native ESC/POS barcode (12-13 digits)
- PDF417: ` ```pdf417 \n data \n ``` ` → Uses ESC/POS PDF417 commands
- DataMatrix: ` ```datamatrix \n data \n ``` ` → Uses ESC/POS DataMatrix commands

*Tables:*
- 2-3 columns → Rendered with ASCII borders, auto-calculated column widths
- 4+ columns → Converted to vertical key-value layout (thermal printer constraint)
- Headers rendered bold and centered
- Text truncated/wrapped to fit column width


**Example Markdown**:
```markdown
# Receipt

## Items

| Item | Price |
|------|-------|
| Coffee | 4.50 |
| Tea | 3.00 |

**Total: EUR 7.50**

```qr
https://pay.example.com/order/12345
```
```

**Test Files**:
- `testdata/simple.md` - Basic formatting (headings, bold, lists)
- `testdata/with_qr.md` - QR code example
- `testdata/checklist.md` - Task lists
- `testdata/with_tables.md` - Table rendering
- `testdata/with_image.md` - Image embedding
- `testdata/wide_table.md` - 4+ column table (vertical layout)
- `testdata/full_example.md` - All features combined

## Development Notes

### When modifying ESC/POS output:
- Always initialize printer first (`0x1B, 0x40`)
- Reset formatting before cutting (`0x1B, 0x40` again)
- Feed paper before cutting (`0x1B, 0x64, 0x05`)
- ESC/POS is stateful - reset alignment/bold after use
- Test on actual hardware - emulators may not match real behavior

### Image Guidelines:
Images should be high contrast and simple. Thermal printers have ~200 DPI resolution and cannot print grayscale. Use the `-threshold` parameter to adjust the darkness of printed images (0-100, higher = more black pixels).

## Dependencies

```go
require (
    github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
    github.com/yuin/goldmark v1.7.8
    golang.org/x/image v0.23.0
)
```

- `golang.org/x/image` - Adds BMP decoder support (via blank import)
- `github.com/skip2/go-qrcode` - QR code generation
- `github.com/yuin/goldmark` - Markdown parsing (CommonMark + GFM support)

## Printer Setup

The application expects a printer configured in macOS CUPS. To set up:
1. Connect printer via USB
2. Add printer in System Preferences → Printers & Scanners
3. Get the name
4. Use -printer argument as cli parameter

Alternatively, use direct serial communication via `/dev/cu.usbserial-*` (see shell scripts).
