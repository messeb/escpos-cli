package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/messeb/escpos-cli/internal/escpos"
	"github.com/messeb/escpos-cli/internal/markdown"
)

func main() {
	printer := flag.String("printer", "Printer_POS_80", "Printer name")
	file := flag.String("file", "", "Markdown file to print")
	output := flag.String("output", "", "Write raw ESC/POS to this file instead of printing (bypasses CUPS)")
	threshold := flag.Int("threshold", 75, "Image threshold (0-100, higher = more black pixels)")
	flag.Parse()

	// Handle positional argument for backward compatibility
	if *file == "" && len(flag.Args()) > 0 {
		*file = flag.Args()[0]
	}

	if *file == "" {
		fmt.Println("Usage: escpos -file <markdown> [-printer <name>] [-output <file>] [-threshold <0-100>]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  escpos -file document.md")
		fmt.Println("  escpos -printer MyPrinter -file document.md")
		fmt.Println("  escpos -file document.md -output printcode_escpos.bin")
		fmt.Println("  escpos -file document.md -threshold 80")
		fmt.Println("  escpos document.md")
		os.Exit(1)
	}

	// Build ESC/POS data from markdown
	data, err := markdown.BuildMarkdownPrint(*file, *threshold)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Export mode: write raw ESC/POS to a file and skip CUPS entirely
	if *output != "" {
		if err := escpos.WriteToFile(data, *output); err != nil {
			fmt.Printf("Output error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Wrote %s\n", *output)
		return
	}

	// Send to printer
	err = escpos.SendToPrinter(data, *printer)
	if err != nil {
		fmt.Printf("Print error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Printed")
}
