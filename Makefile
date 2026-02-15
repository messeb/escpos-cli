.PHONY: build clean test lint lint-fix fmt vet install help coverage-check

# Binary name
BINARY=escpos

# Build directory
BUILD_DIR=./dist

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Main package path
MAIN_PATH=./cmd/escpos

# Default target
all: build

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build:
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY) $(MAIN_PATH)

## build-all: Build for multiple platforms
build-all:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 $(MAIN_PATH)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY)-linux-arm64 $(MAIN_PATH)

## install: Install the binary to GOPATH/bin
install:
	$(GOCMD) install $(MAIN_PATH)

## clean: Remove build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY)
	rm -f $(BUILD_DIR)/$(BINARY)-*
	rm -f /tmp/print_*.bin

## test: Run tests
test:
	$(GOTEST) -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## coverage-check: Check if coverage meets threshold (80%)
coverage-check:
	@echo "Checking code coverage..."
	@if [ -f coverage.out ]; then \
		total_line=$$(go tool cover -func=coverage.out | grep total); \
		if [ -z "$$total_line" ]; then \
			echo "⚠️  No test files found - skipping coverage check"; \
			exit 0; \
		fi; \
		coverage=$$(echo "$$total_line" | awk '{print $$3}' | sed 's/%//'); \
		threshold=80; \
		echo "Current coverage: $$coverage%"; \
		echo "Required threshold: $$threshold%"; \
		if [ $$(echo "$$coverage < $$threshold" | bc -l) -eq 1 ]; then \
			echo "⚠️  Coverage $$coverage% is below threshold $$threshold%"; \
			echo "ℹ️  Add tests to improve coverage"; \
			exit 0; \
		else \
			echo "✓ Coverage $$coverage% meets threshold"; \
		fi \
	else \
		echo "❌ coverage.out not found. Run 'make test-coverage' first"; \
		exit 1; \
	fi

## fmt: Format Go code
fmt:
	$(GOFMT) ./...

## vet: Run go vet
vet:
	$(GOVET) ./...

## lint: Run golangci-lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  brew install golangci-lint  (macOS)"; \
		echo "  or visit https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

## lint-fix: Run golangci-lint and auto-fix issues
lint-fix:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix ./...; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  brew install golangci-lint  (macOS)"; \
		echo "  or visit https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

## tidy: Tidy go.mod
tidy:
	$(GOMOD) tidy

## deps: Download dependencies
deps:
	$(GOMOD) download

## check: Run fmt, vet, and lint
check: fmt vet lint

## run: Build and run with example file
run: build
	$(BUILD_DIR)/$(BINARY) -file testdata/simple.md

## print-test: Print all test files
print-test: build
	@echo "Printing test files..."
	$(BUILD_DIR)/$(BINARY) -file testdata/simple.md
	@sleep 2
	$(BUILD_DIR)/$(BINARY) -file testdata/with_qr.md
	@sleep 2
	$(BUILD_DIR)/$(BINARY) -file testdata/line_breaks_test.md
