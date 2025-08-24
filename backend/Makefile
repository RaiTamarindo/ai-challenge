# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=api
BINARY_PATH=./bin/$(BINARY_NAME)

# Docker parameters
DOCKER_IMAGE=feature-voting-backend

.PHONY: build run clean test test-verbose test-coverage deps docker-build docker-run migrate-up migrate-down generate-mocks

# Build the application
build:
	$(GOBUILD) -o $(BINARY_PATH) ./cmd/api

# Run the application
run:
	$(GOCMD) run ./cmd/api

# Clean build files
clean:
	$(GOCLEAN)
	rm -rf ./bin

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with verbose output and race detection
test-verbose:
	$(GOTEST) -v -race ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Download dependencies
deps:
	$(GOMOD) tidy
	$(GOMOD) download

# Docker build
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Docker run with environment variables
docker-run:
	docker run -d \
		--name feature-voting-backend \
		-p 8080:8080 \
		-e DATABASE_URL=$(DATABASE_URL) \
		-e JWT_SECRET=$(JWT_SECRET) \
		$(DOCKER_IMAGE)

# Database migrations up
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

# Database migrations down
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

# Create new migration
migration:
	migrate create -ext sql -dir migrations -seq $(name)

# Generate mocks using mockery
generate-mocks:
	@echo "Generating mocks..."
	$$HOME/go/bin/mockery --config .mockery.yaml
	@echo "Mocks generated successfully!"

# Development helpers
dev: deps generate-mocks test build

# CI/CD helpers
ci: deps test-coverage

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application binary"
	@echo "  run           - Run the application directly"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run all tests"
	@echo "  test-verbose  - Run tests with verbose output and race detection"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  migrate-up    - Run database migrations up"
	@echo "  migrate-down  - Run database migrations down"
	@echo "  generate-mocks - Generate mocks using mockery"
	@echo "  dev           - Full development setup (deps, mocks, test, build)"
	@echo "  ci            - CI pipeline (deps, test with coverage)"
	@echo "  help          - Show this help message"