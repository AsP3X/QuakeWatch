# Makefile for QuakeWatch Scraper
.PHONY: build clean test run install fmt lint docs

# Build the application
build:
	go build -o bin/quakewatch-scraper cmd/scraper/main.go

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/quakewatch-scraper-linux-amd64 cmd/scraper/main.go

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/quakewatch-scraper-darwin-amd64 cmd/scraper/main.go

build-windows:
	GOOS=windows GOARCH=amd64 go build -o bin/quakewatch-scraper-windows-amd64.exe cmd/scraper/main.go

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Run tests
test:
	go test ./...

# Run the application
run: build
	./bin/quakewatch-scraper

# Install dependencies
install:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate documentation
docs:
	godoc -http=:6060

# Create necessary directories
setup:
	mkdir -p bin data/earthquakes data/faults

# Quick test of the application
test-app: build
	./bin/quakewatch-scraper version
	./bin/quakewatch-scraper health

# Collect recent earthquakes (test)
test-earthquakes: build
	./bin/quakewatch-scraper earthquakes recent --limit 10

# Collect fault data (test)
test-faults: build
	./bin/quakewatch-scraper faults collect

# Test purge command (dry run)
test-purge: build
	./bin/quakewatch-scraper purge --dry-run 