# go-finding Justfile

# Default recipe
[private]
default:
    @just --list

# Run all tests
test:
    go test -v -race ./...

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Run with coverage
cover:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
    golangci-lint run ./...

# Format code
fmt:
    go fmt ./...
    goimports -w .

# Download dependencies
deps:
    go mod download
    go mod tidy

# Clean build artifacts
clean:
    rm -f coverage.out coverage.html
    go clean -cache

# Build (library, no binary)
build:
    go build ./...

# Check for vulnerabilities
vuln:
    govulncheck ./...

# Run all checks
check: fmt lint test
    @echo "All checks passed!"

# Update dependencies
update:
    go get -u ./...
    go mod tidy

# Generate documentation
docs:
    go doc -all . > API.md

# CI simulation
ci: deps check
    @echo "CI checks complete"
