# ESC/POS Package Tests

## Quick Start

```bash
# Run all tests
go test ./internal/escpos/...

# Run with coverage
go test ./internal/escpos/... -cover

# Run with detailed coverage report
go test ./internal/escpos/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run benchmarks
go test ./internal/escpos/... -bench=. -benchtime=1s

# Run specific test
go test ./internal/escpos/... -run TestGenerateQRCode -v
```

## Test Statistics

- **Total Tests:** 33 test functions with 62+ subtests
- **Coverage:** 97.0% of statements
- **Test Files:** 4 files (commands_test.go, barcode_test.go, image_test.go, printer_test.go)
- **Benchmarks:** 16 benchmark functions

## Test Files Overview

### commands_test.go
Tests ESC/POS command generation functions.

**Functions Tested:**
- Init, LineFeed, Feed, Cut
- AlignCenter, AlignLeft
- BoldOn, BoldOff
- UnderlineOn, UnderlineOff
- DoubleSize, DoubleHeight, DoubleWidth, NormalSize
- SolidLine

**Coverage:** 100%

### barcode_test.go
Tests barcode and QR code generation.

**Functions Tested:**
- GenerateQRCode
- GeneratePDF417
- GenerateDataMatrix
- GenerateEAN13

**Coverage:** 100%

### image_test.go
Tests image conversion to ESC/POS bitmap format.

**Functions Tested:**
- ConvertImageToESCPOS

**Test Features:**
- Automatic image scaling (max 384px width)
- Threshold testing (0-100%)
- Error handling for invalid files
- Edge cases (1x1, very wide, very tall images)

**Coverage:** 100%

### printer_test.go
Tests printer job management and data transmission.

**Functions Tested:**
- SendToPrinter

**Test Features:**
- File creation and cleanup
- Concurrent print jobs
- Data integrity verification
- Printer name validation
- Integration test with full print job

**Coverage:** 66.7% (limited by external lpr command)

## Test Design

All tests follow Go best practices:

1. **Table-Driven Tests** - Comprehensive test cases in readable format
2. **Subtests** - Organized test execution with clear naming
3. **Byte-Level Validation** - Exact ESC/POS protocol verification
4. **Error Testing** - Invalid inputs and edge cases covered
5. **Benchmarks** - Performance baselines for all major functions
6. **Helper Functions** - Reusable test utilities (image creation, etc.)
7. **CI-Friendly** - Printer tests skip in CI environments

## Running Specific Test Categories

```bash
# Command tests only
go test ./internal/escpos/... -run TestInit

# Barcode tests only
go test ./internal/escpos/... -run TestGenerate

# Image tests only
go test ./internal/escpos/... -run TestConvert

# Printer tests only
go test ./internal/escpos/... -run TestSendToPrinter

# Benchmark QR code generation
go test ./internal/escpos/... -bench=BenchmarkGenerateQRCode

# Benchmark image conversion
go test ./internal/escpos/... -bench=BenchmarkConvertImage
```

## CI/CD Integration

The test suite is designed to run in CI/CD environments:

```yaml
# Example GitHub Actions workflow
- name: Run tests
  run: go test ./internal/escpos/... -v -cover

- name: Generate coverage report
  run: |
    go test ./internal/escpos/... -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html
```

Printer tests automatically skip when `CI` environment variable is set.

## Coverage Details

```
Total Coverage: 97.0%

commands.go:    100.0%  - All ESC/POS command functions
barcode.go:     100.0%  - QR, PDF417, DataMatrix, EAN-13
image.go:       100.0%  - Image conversion to bitmap
printer.go:      66.7%  - Print job management
```

The printer.go coverage is limited because:
- External system commands (lpr, cancel) cannot be fully mocked
- Physical printer hardware not available in test environment
- File I/O cleanup happens in goroutines

## Example Test Execution

```
$ go test ./internal/escpos/... -v

=== RUN   TestInit
--- PASS: TestInit (0.00s)

=== RUN   TestGenerateQRCode
=== RUN   TestGenerateQRCode/simple_text
=== RUN   TestGenerateQRCode/URL
--- PASS: TestGenerateQRCode (0.00s)

=== RUN   TestConvertImageToESCPOS
=== RUN   TestConvertImageToESCPOS/small_image
=== RUN   TestConvertImageToESCPOS/medium_image
--- PASS: TestConvertImageToESCPOS (0.01s)

PASS
ok  	github.com/messeb/escpos-cli/internal/escpos	4.600s
```

## Benchmark Results (Apple M4 Pro)

```
BenchmarkGenerateQRCode-14                    3642    323801 ns/op
BenchmarkGeneratePDF417-14                17924157        65.26 ns/op
BenchmarkGenerateDataMatrix-14            26748099        45.34 ns/op
BenchmarkGenerateEAN13-14                 27116677        44.93 ns/op
BenchmarkConvertImageToESCPOS_Small-14        8473    142980 ns/op
BenchmarkConvertImageToESCPOS_Large-14         562   2130574 ns/op
```

## Troubleshooting

**Tests fail with "printer not found"**
- This is expected behavior - tests verify error handling
- Printer tests are designed to work without physical hardware
- Check that tests still pass despite printer errors

**Coverage report shows missing lines**
- Some cleanup code runs in deferred functions
- External command errors may not be reached
- This is acceptable for system integration code

**Benchmarks are slow**
- QR code generation involves image processing
- Image conversion scales with size
- This is expected behavior

## Adding New Tests

When adding new functionality to the escpos package:

1. Create tests in the corresponding test file
2. Follow table-driven test pattern
3. Test both success and error cases
4. Add benchmarks for performance-critical code
5. Validate exact byte sequences for ESC/POS commands
6. Run `go test -cover` to verify coverage stays >80%

Example test structure:

```go
func TestNewFunction(t *testing.T) {
    tests := []struct {
        name string
        input string
        want []byte
    }{
        {
            name: "basic case",
            input: "test",
            want: []byte{0x1B, 0x40},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := NewFunction(tt.input)
            if !bytes.Equal(got, tt.want) {
                t.Errorf("NewFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Additional Resources

- [TEST_COVERAGE.md](TEST_COVERAGE.md) - Detailed coverage report
- [Go Testing Package](https://pkg.go.dev/testing) - Official documentation
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests) - Best practices
