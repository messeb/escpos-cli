package escpos

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// WriteToFile writes raw ESC/POS data to the given path. This is the export
// path used instead of SendToPrinter when CUPS/lpr is unavailable (e.g. Windows).
func WriteToFile(data []byte, path string) error {
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write output file %q: %w", path, err)
	}
	return nil
}

// SendToPrinter sends ESC/POS data to the specified printer via lpr
func SendToPrinter(data []byte, printerName string) error {
	// Clear any stuck jobs
	_ = exec.Command("cancel", "-a").Run()
	time.Sleep(200 * time.Millisecond)

	// Write to temporary file
	tmpFile := fmt.Sprintf("/tmp/print_%d.bin", time.Now().Unix())
	err := os.WriteFile(tmpFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Print via lpr
	cmd := exec.Command("lpr", "-P", printerName, "-o", "raw", tmpFile)
	err = cmd.Run()
	if err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("print failed: %w", err)
	}

	// Cleanup
	time.Sleep(100 * time.Millisecond)
	_ = os.Remove(tmpFile)
	_ = exec.Command("cancel", "-a").Run()

	return nil
}
