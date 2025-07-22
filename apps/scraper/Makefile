# Makefile for QuakeWatch Scraper
.PHONY: build clean test run install fmt docs

# Build the application
build:
	go build -o bin/quakewatch-scraper cmd/scraper/main.go

# Build for multiple platforms
build-all: build-linux build-darwin

build-linux:
	GOOS=linux GOARCH=amd64 go build -o bin/quakewatch-scraper-linux-amd64 cmd/scraper/main.go

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o bin/quakewatch-scraper-darwin-amd64 cmd/scraper/main.go



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

# Database commands
db-migrate-up: build
	./bin/quakewatch-scraper db migrate up

db-migrate-down: build
	./bin/quakewatch-scraper db migrate down

db-status: build
	./bin/quakewatch-scraper db status

db-init: build
	./bin/quakewatch-scraper db init

# Database setup with Docker (requires Docker)
db-setup-docker:
	docker run --name quakewatch-postgres \
		-e POSTGRES_DB=quakewatch \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-p 5432:5432 \
		-d postgres:15

db-stop-docker:
	docker stop quakewatch-postgres || true
	docker rm quakewatch-postgres || true

# Environment setup
setup-env:
	@echo "Setting up environment variables..."
	@echo "DB_HOST=localhost" > .env
	@echo "DB_PORT=5432" >> .env
	@echo "DB_USER=postgres" >> .env
	@echo "DB_PASSWORD=postgres" >> .env
	@echo "DB_NAME=quakewatch" >> .env
	@echo "DB_SSL_MODE=disable" >> .env
	@echo "Environment file created: .env" 