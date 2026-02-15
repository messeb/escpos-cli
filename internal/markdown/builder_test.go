package markdown

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/messeb/escpos-cli/internal/escpos"
)

func TestBuildMarkdownPrint(t *testing.T) {
	tests := []struct {
		name           string
		fileName       string
		threshold      int
		wantErr        bool
		checkContains  [][]byte // Check if output contains these byte sequences
		checkNotContains [][]byte // Check if output does NOT contain these sequences
	}{
		{
			name:      "simple markdown with headings and lists",
			fileName:  "simple.md",
			threshold: 75,
			wantErr:   false,
			checkContains: [][]byte{
				escpos.Init(),
				escpos.AlignCenter(),
				escpos.BoldOn(),
				escpos.DoubleSize(), // H1
				[]byte("Simple Test"),
				escpos.AlignLeft(),
				[]byte("Features"),
			},
		},
		{
			name:      "checklist with task items",
			fileName:  "checklist.md",
			threshold: 75,
			wantErr:   false,
			checkContains: [][]byte{
				[]byte("Todo List"),
				[]byte("[x]"), // Checked item
				[]byte("[ ]"), // Unchecked item
				[]byte("Wake up"),
				[]byte("Go to work"),
			},
		},
		{
			name:      "table rendering",
			fileName:  "with_tables.md",
			threshold: 75,
			wantErr:   false,
			checkContains: [][]byte{
				[]byte("Receipt Example"),
				[]byte("Item"),
				[]byte("Qty"),
				[]byte("Price"),
				[]byte("Coffee"),
			},
		},
		{
			name:      "non-existent file",
			fileName:  "does_not_exist.md",
			threshold: 75,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build absolute path to test file
			testDataPath := filepath.Join("..", "..", "testdata", tt.fileName)

			// Check if file exists (for negative tests)
			if tt.wantErr {
				_, err := os.Stat(testDataPath)
				if err == nil {
					// File exists, create a path that doesn't exist
					testDataPath = filepath.Join("..", "..", "testdata", "this_file_does_not_exist.md")
				}
			}

			got, err := BuildMarkdownPrint(testDataPath, tt.threshold)

			if (err != nil) != tt.wantErr {
				t.Errorf("BuildMarkdownPrint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check that output is not empty
			if len(got) == 0 {
				t.Error("BuildMarkdownPrint() returned empty output")
			}

			// Check for required byte sequences
			for _, check := range tt.checkContains {
				if !containsBytes(got, check) {
					t.Errorf("BuildMarkdownPrint() output does not contain expected bytes: %v", check)
				}
			}

			// Check that certain sequences are NOT present
			for _, check := range tt.checkNotContains {
				if containsBytes(got, check) {
					t.Errorf("BuildMarkdownPrint() output contains unexpected bytes: %v", check)
				}
			}

			// Check that output starts with Init and ends with Feed + Cut
			if !startsWithBytes(got, escpos.Init()) {
				t.Error("BuildMarkdownPrint() output does not start with Init()")
			}

			if !endsWithBytes(got, escpos.Cut()) {
				t.Error("BuildMarkdownPrint() output does not end with Cut()")
			}
		})
	}
}

func TestBuildMarkdownPrintIntegration(t *testing.T) {
	// Integration test that processes actual markdown files
	testFiles := []string{
		"simple.md",
		"checklist.md",
		"with_tables.md",
	}

	for _, fileName := range testFiles {
		t.Run(fileName, func(t *testing.T) {
			testDataPath := filepath.Join("..", "..", "testdata", fileName)

			// Skip if file doesn't exist
			if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
				t.Skipf("Test file %s does not exist", testDataPath)
				return
			}

			output, err := BuildMarkdownPrint(testDataPath, 75)
			if err != nil {
				t.Fatalf("BuildMarkdownPrint() failed: %v", err)
			}

			// Validate output structure
			if len(output) < 10 {
				t.Error("Output is suspiciously short")
			}

			// Check for ESC/POS structure
			if !containsBytes(output, escpos.Init()) {
				t.Error("Output missing Init command")
			}

			if !containsBytes(output, escpos.Cut()) {
				t.Error("Output missing Cut command")
			}
		})
	}
}

func TestBufWriter(t *testing.T) {
	var buf bytes.Buffer
	bw := &bufWriter{buf: &buf}

	// Test Write
	n, err := bw.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != 4 {
		t.Errorf("Write() wrote %d bytes, want 4", n)
	}

	// Test WriteByte
	err = bw.WriteByte('x')
	if err != nil {
		t.Errorf("WriteByte() error = %v", err)
	}

	// Test WriteString
	n, err = bw.WriteString("hello")
	if err != nil {
		t.Errorf("WriteString() error = %v", err)
	}
	if n != 5 {
		t.Errorf("WriteString() wrote %d bytes, want 5", n)
	}

	// Test WriteRune
	_, err = bw.WriteRune('!')
	if err != nil {
		t.Errorf("WriteRune() error = %v", err)
	}

	// Test Flush (should be no-op)
	err = bw.Flush()
	if err != nil {
		t.Errorf("Flush() error = %v", err)
	}

	// Verify content
	got := string(bw.Bytes())
	want := "testxhello!"
	if got != want {
		t.Errorf("bufWriter content = %q, want %q", got, want)
	}

	// Test Available
	available := bw.Available()
	if available < 0 {
		t.Errorf("Available() = %d, should be non-negative", available)
	}

	// Test Buffered
	buffered := bw.Buffered()
	if buffered != len(want) {
		t.Errorf("Buffered() = %d, want %d", buffered, len(want))
	}
}

// Helper functions

func containsBytes(haystack, needle []byte) bool {
	if len(needle) == 0 {
		return true
	}
	if len(haystack) < len(needle) {
		return false
	}

	for i := 0; i <= len(haystack)-len(needle); i++ {
		found := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				found = false
				break
			}
		}
		if found {
			return true
		}
	}
	return false
}

func startsWithBytes(data, prefix []byte) bool {
	if len(data) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if data[i] != prefix[i] {
			return false
		}
	}
	return true
}

func endsWithBytes(data, suffix []byte) bool {
	if len(data) < len(suffix) {
		return false
	}
	offset := len(data) - len(suffix)
	for i := 0; i < len(suffix); i++ {
		if data[offset+i] != suffix[i] {
			return false
		}
	}
	return true
}
