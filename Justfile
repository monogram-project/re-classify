# Justfile for re-classify

# Variables
binary_name := "re-classify"
cmd_dir := "./cmd/re-classify"
version := `git describe --tags --always --dirty 2>/dev/null || echo "unknown"`
ldflags := "-ldflags \"-X main.Version=" + version + "\""

# Show available recipes (default)
default: help

# Build the binary
build:
    go build {{ldflags}} -o {{binary_name}} {{cmd_dir}}

# Install the binary to GOBIN/GOPATH
install:
    go install {{ldflags}} {{cmd_dir}}

# Run tests
test:
    go test -v ./...

# Clean build artifacts
clean:
    go clean
    rm -f {{binary_name}}

# Format code
fmt:
    go fmt ./...

# Vet code
vet:
    go vet ./...

# Lint code (requires golangci-lint)
lint:
    golangci-lint run

# Run all checks (fmt, vet, test)
check: fmt vet test

# Development build with version info
dev-build:
    go build {{ldflags}} -o {{binary_name}} {{cmd_dir}}

# Show available recipes
help:
    @just --list
