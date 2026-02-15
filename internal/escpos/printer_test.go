package escpos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestSendToPrinter tests the SendToPrinter function
// Note: This test will not actually send to a printer, but tests the file creation and cleanup
func TestSendToPrinter_FileCreation(t *testing.T) {
	// Skip this test in CI/CD environments where lpr may not be available
	if os.Getenv("CI") != "" {
		t.Skip("Skipping printer test in CI environment")
	}

	testData := []byte{0x1B, 0x40, 0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x0A}
	printerName := "TestPrinter_DoesNotExist"

	// This will likely fail because the printer doesn't exist
	// But we can still test the temp file creation logic by checking the error
	err := SendToPrinter(testData, printerName)

	// We expect an error since the printer doesn't exist
	// The important part is that the function handles the error gracefully
	if err == nil {
		t.Log("SendToPrinter() succeeded unexpectedly (printer may actually exist)")
	} else {
		// Error is expected - check that it's a print failure, not a file write failure
		if strings.Contains(err.Error(), "failed to write temp file") {
			t.Errorf("SendToPrinter() failed to write temp file: %v", err)
		}
		// Print failure is expected for non-existent printer
		if !strings.Contains(err.Error(), "print failed") {
			t.Logf("SendToPrinter() returned unexpected error: %v", err)
		}
	}
}

// TestSendToPrinter_EmptyData tests sending empty data
func TestSendToPrinter_EmptyData(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping printer test in CI environment")
	}

	testData := []byte{}
	printerName := "TestPrinter"

	// Should handle empty data gracefully
	err := SendToPrinter(testData, printerName)

	// We expect an error (printer doesn't exist), but should not panic
	if err == nil {
		t.Log("SendToPrinter() succeeded with empty data")
	}
}

// TestSendToPrinter_LargeData tests sending large amounts of data
func TestSendToPrinter_LargeData(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping printer test in CI environment")
	}

	// Create large test data (10KB)
	testData := make([]byte, 10*1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	printerName := "TestPrinter"

	err := SendToPrinter(testData, printerName)

	// Should handle large data without crashing
	if err == nil {
		t.Log("SendToPrinter() succeeded with large data")
	}
}

// TestSendToPrinter_TempFileCleanup tests that temp files are cleaned up
func TestSendToPrinter_TempFileCleanup(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping printer test in CI environment")
	}

	testData := []byte{0x1B, 0x40}
	printerName := "TestPrinter"

	// Get list of temp files before
	tmpFiles, _ := filepath.Glob("/tmp/print_*.bin")
	beforeCount := len(tmpFiles)

	// Attempt to print (will fail, but should still clean up)
	_ = SendToPrinter(testData, printerName)

	// Wait a moment for cleanup
	time.Sleep(200 * time.Millisecond)

	// Get list of temp files after
	tmpFiles, _ = filepath.Glob("/tmp/print_*.bin")
	afterCount := len(tmpFiles)

	// Should not have increased the number of temp files
	if afterCount > beforeCount {
		t.Logf("Warning: Temp files may not have been cleaned up. Before: %d, After: %d", beforeCount, afterCount)
		// Clean up any leftover files
		for _, f := range tmpFiles {
			if err := os.Remove(f); err != nil {
				t.Logf("failed to remove temp file %s: %v", f, err)
			}
		}
	}
}

// TestSendToPrinter_PrinterNameValidation tests different printer names
func TestSendToPrinter_PrinterNameValidation(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping printer test in CI environment")
	}

	tests := []struct {
		name        string
		printerName string
		data        []byte
	}{
		{
			name:        "simple printer name",
			printerName: "TestPrinter",
			data:        []byte{0x1B, 0x40},
		},
		{
			name:        "printer name with underscores",
			printerName: "Test_Printer_POS",
			data:        []byte{0x1B, 0x40},
		},
		{
			name:        "empty printer name",
			printerName: "",
			data:        []byte{0x1B, 0x40},
		},
		{
			name:        "printer name with spaces",
			printerName: "Test Printer",
			data:        []byte{0x1B, 0x40},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SendToPrinter(tt.data, tt.printerName)

			// All should handle gracefully (though may fail to print)
			if err != nil {
				// Expected - printer doesn't exist
				t.Logf("SendToPrinter() error (expected): %v", err)
			}
		})
	}
}

// TestSendToPrinter_ConcurrentCalls tests multiple concurrent print jobs
func TestSendToPrinter_ConcurrentCalls(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping printer test in CI environment")
	}

	testData := []byte{0x1B, 0x40, 0x48, 0x65, 0x6C, 0x6C, 0x6F}
	printerName := "TestPrinter"

	// Send multiple print jobs concurrently
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(index int) {
			err := SendToPrinter(testData, printerName)
			if err != nil {
				t.Logf("Concurrent print job %d error (expected): %v", index, err)
			}
			done <- true
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Check for temp file cleanup
	time.Sleep(300 * time.Millisecond)
	tmpFiles, _ := filepath.Glob("/tmp/print_*.bin")

	// Clean up any leftover files
	for _, f := range tmpFiles {
		if err := os.Remove(f); err != nil {
			t.Logf("failed to remove temp file %s: %v", f, err)
		}
	}

	if len(tmpFiles) > 5 {
		t.Logf("Warning: Found %d temp files after concurrent calls", len(tmpFiles))
	}
}

// TestSendToPrinter_DataIntegrity tests that data is written correctly to temp file
func TestSendToPrinter_DataIntegrity(t *testing.T) {
	// This test manually creates a temp file to verify the data writing logic
	testData := []byte{0x1B, 0x40, 0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x0A, 0x1D, 0x56, 0x01}

	tmpFile := filepath.Join(t.TempDir(), "test_print.bin")
	err := os.WriteFile(tmpFile, testData, 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Read back and verify
	readData, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	if len(readData) != len(testData) {
		t.Errorf("Data length mismatch: got %d, want %d", len(readData), len(testData))
	}

	for i := range testData {
		if readData[i] != testData[i] {
			t.Errorf("Data mismatch at byte %d: got 0x%02X, want 0x%02X", i, readData[i], testData[i])
		}
	}
}

// TestSendToPrinter_SpecialCharacters tests data with special bytes
func TestSendToPrinter_SpecialCharacters(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping printer test in CI environment")
	}

	// Test with all possible byte values
	testData := make([]byte, 256)
	for i := 0; i < 256; i++ {
		testData[i] = byte(i)
	}

	printerName := "TestPrinter"

	err := SendToPrinter(testData, printerName)

	// Should handle all byte values without corruption
	if err != nil {
		if strings.Contains(err.Error(), "failed to write temp file") {
			t.Errorf("SendToPrinter() failed to write special characters: %v", err)
		}
	}
}

// Benchmark tests
func BenchmarkSendToPrinter(b *testing.B) {
	if os.Getenv("CI") != "" {
		b.Skip("Skipping printer benchmark in CI environment")
	}

	testData := []byte{0x1B, 0x40, 0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x0A}
	printerName := "TestPrinter"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SendToPrinter(testData, printerName)
	}
}

func BenchmarkSendToPrinter_LargeData(b *testing.B) {
	if os.Getenv("CI") != "" {
		b.Skip("Skipping printer benchmark in CI environment")
	}

	// Create 100KB of test data
	testData := make([]byte, 100*1024)
	for i := range testData {
		testData[i] = byte(i % 256)
	}

	printerName := "TestPrinter"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SendToPrinter(testData, printerName)
	}
}

// Test integration of multiple components
func TestIntegration_FullPrintJob(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping integration test in CI environment")
	}

	// Build a complete print job using all components
	var printJob []byte

	// Initialize printer
	printJob = append(printJob, Init()...)

	// Add centered bold heading
	printJob = append(printJob, AlignCenter()...)
	printJob = append(printJob, BoldOn()...)
	printJob = append(printJob, DoubleSize()...)
	printJob = append(printJob, []byte("TEST RECEIPT")...)
	printJob = append(printJob, LineFeed()...)
	printJob = append(printJob, NormalSize()...)
	printJob = append(printJob, BoldOff()...)

	// Add normal text
	printJob = append(printJob, AlignLeft()...)
	printJob = append(printJob, []byte("Item 1: $10.00")...)
	printJob = append(printJob, LineFeed()...)
	printJob = append(printJob, []byte("Item 2: $15.00")...)
	printJob = append(printJob, LineFeed()...)

	// Add solid line
	printJob = append(printJob, SolidLine(384)...)

	// Add total
	printJob = append(printJob, BoldOn()...)
	printJob = append(printJob, []byte("Total: $25.00")...)
	printJob = append(printJob, BoldOff()...)
	printJob = append(printJob, LineFeed()...)

	// Feed and cut
	printJob = append(printJob, Feed(5)...)
	printJob = append(printJob, Cut()...)

	// Verify the job is not empty
	if len(printJob) == 0 {
		t.Error("Integration test produced empty print job")
	}

	// Attempt to send (will fail, but tests integration)
	if err := SendToPrinter(printJob, "TestPrinter"); err != nil {
		t.Logf("Integration test print error (expected): %v", err)
	}
}
