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
	threshold := flag.Int("threshold", 75, "Image threshold (0-100, higher = more black pixels)")
	flag.Parse()

	// Handle positional argument for backward compatibility
	if *file == "" && len(flag.Args()) > 0 {
		*file = flag.Args()[0]
	}

	if *file == "" {
		fmt.Println("Usage: escpos -file <markdown> [-printer <name>] [-threshold <0-100>]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  escpos -file document.md")
		fmt.Println("  escpos -printer MyPrinter -file document.md")
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

	// Send to printer
	err = escpos.SendToPrinter(data, *printer)
	if err != nil {
		fmt.Printf("Print error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Printed")
}
