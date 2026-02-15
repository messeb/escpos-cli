package escpos

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

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
