# Design: `-output` raw file export

**Date:** 2026-07-03
**Branch:** `feature/output-raw-file`

## Motivation

The CLI only sends ESC/POS output to a printer via macOS CUPS (`lpr`). Windows
users (and anyone who wants to capture the raw byte stream) have no way to get
the generated ESC/POS bytes, because CUPS is macOS/Linux-only. A user building
receipt-generation tooling on Windows needs a plain file of raw ESC/POS codes
that they can send to a printer through their own transport.

## Goal

Add an `-output <file>` flag that writes the raw ESC/POS byte array to a file
and bypasses the CUPS/`lpr` path entirely.

## Behavior

- `-output <file>` set: build the ESC/POS bytes from the markdown, write them
  verbatim to `<file>`, print `✓ Wrote <file>`, and exit. `SendToPrinter`,
  `lpr`, and `cancel` are never invoked, so this works on Windows.
- `-output` empty (default): behavior unchanged — print via CUPS.
- `-printer` is ignored when `-output` is set (not an error). This keeps
  invocations simple; there is no meaningful printer in export mode.

Example:

```bash
escpos -file receipt.md -output printcode_escpos.bin
```

## Code Changes

### `internal/escpos/printer.go`

Add:

```go
// WriteToFile writes raw ESC/POS data to the given path.
func WriteToFile(data []byte, path string) error
```

A thin `os.WriteFile` wrapper (mode `0644`) that wraps any error with context.
Lives in the `escpos` package alongside `SendToPrinter` so all output
transports stay together.

### `cmd/escpos/main.go`

- Register `output := flag.String("output", "", "Write raw ESC/POS to file instead of printing")`.
- After `BuildMarkdownPrint`, branch:
  - `*output != ""` → `escpos.WriteToFile(data, *output)`; on success print
    `✓ Wrote <file>`.
  - else → existing `escpos.SendToPrinter` path.
- Update the usage text and examples to include `-output`.

## Testing

`internal/escpos/printer_test.go`, table-driven test for `WriteToFile`:

- writes the exact bytes to a temp file (read back, compare).
- returns an error for an unwritable path (e.g. a nonexistent directory).

No hardware or CUPS dependency. `SendToPrinter` remains untested (unchanged,
requires a real printer).

## Out of Scope (YAGNI)

- Subcommand restructuring (`print` / `export`).
- Writing to stdout / streaming.
- Simultaneous file + print.
