# Makefile for go-user-service

# Define variables
SERVICE_NAME = go-user-service
BUILD_DIR = ./bin
CMD_DIR = ./cmd/server
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# --- Build and Run ---

# Build the service with version information
build: 
	@echo "Building $(SERVICE_NAME) version $(VERSION)..."
	go build -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)" -o $(BUILD_DIR)/$(SERVICE_NAME) $(CMD_DIR)
	@echo "Build complete."

# Run the service
run: build
	@echo "Running $(SERVICE_NAME)..."
	$(BUILD_DIR)/$(SERVICE_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete."

# --- Testing and Code Quality ---

# Run tests
test:
	@echo "Running tests..."
	go test ./...
	@echo "Tests complete."

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

# Run linter
lint:
	@echo "Linting code..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Vetting code..."
	go vet ./...

# Generate mocks for testing
mocks:
	@echo "Generating mocks..."
	mockgen -source=internal/domain/user/repository.go -destination=internal/mocks/user_repository_mock.go -package=mocks
	mockgen -source=internal/domain/auth/repository.go -destination=internal/mocks/auth_repository_mock.go -package=mocks
	mockgen -source=internal/domain/user/service.go -destination=internal/mocks/user_service_mock.go -package=mocks
	mockgen -source=internal/domain/auth/service.go -destination=internal/mocks/auth_service_mock.go -package=mocks
	@echo "Mock generation complete."

# --- Dependency Injection ---

# Regenerate wire dependency injection code
wire:
	@echo "Regenerating wire dependency injection code..."
	cd $(CMD_DIR)/wire && wire
	@echo "Wire code generation complete."

# --- Protobuf Generation ---

PROTO_DIR = ./api/proto
PROTO_OUT = ./internal/transport/grpc
SWAGGER_OUT = ./api/swagger

# Install protoc and required plugins
proto-install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Clean generated protobuf files
proto-clean:
	@echo "Cleaning generated protobuf Go files from api/proto..."
	find $(PROTO_DIR) -name '*.pb.go' -type f -delete
	find $(PROTO_DIR) -name '*.pb.gw.go' -type f -delete
	find $(PROTO_DIR) -name '*_grpc.pb.go' -type f -delete
	@echo "Cleaning generated swagger files from docs/swagger..."
	rm -rf ./docs/swagger/*

# Generate protobuf code and swagger docs
proto-gen: proto-clean
	cd $(PROTO_DIR) && buf generate

# Generate swagger docs
proto-swagger: proto-gen
	@echo "Swagger documentation generated at $(SWAGGER_OUT)/user_service.swagger.json"

# --- Docker Support ---

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(SERVICE_NAME):$(VERSION) .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 8081:8081 $(SERVICE_NAME):$(VERSION)

# --- Database Migration --- 

MIGRATE_CLI = migrate
MIGRATE_DIR = ./migrations
# Use DATABASE_URL environment variable or default to PostgreSQL connection string
# For migrate tool, the PostgreSQL driver should be specified as 'postgres'
DB_URL = $(if $(DATABASE_URL),$(DATABASE_URL),postgres://ewu:123456@localhost:5432/user_auth_dev?sslmode=disable)

# Version to force when fixing dirty migrations
MIGRATE_VERSION = 0

migrate-create:
	@echo "Creating migration file..."
	$(MIGRATE_CLI) create -ext sql -dir $(MIGRATE_DIR) $(name)

migrate-up:
	@echo "Running migrations up..."
	$(MIGRATE_CLI) -database $(DB_URL) -path $(MIGRATE_DIR) up

migrate-down:
	@echo "Running migrations down..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "DATABASE_URL environment variable is not set. Using default connection string."; \
		echo "If this fails, please set the DATABASE_URL environment variable with the correct connection string."; \
	fi
	$(MIGRATE_CLI) -database $(DB_URL) -path $(MIGRATE_DIR) down -all

migrate-force:
	@echo "Forcing migration version to $(MIGRATE_VERSION)..."
	$(MIGRATE_CLI) -database $(DB_URL) -path $(MIGRATE_DIR) force $(MIGRATE_VERSION)

# --- Development Setup ---

# Install development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(MAKE) proto-install
	@echo "Development dependencies installed."

# --- Help ---

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the service"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  clean          - Clean build artifacts"
	@echo "  run            - Build and run the service"
	@echo "  wire           - Regenerate wire dependency injection code"
	@echo "  proto-install  - Install protobuf tools"
	@echo "  proto-gen      - Generate protobuf code"
	@echo "  proto-swagger  - Generate swagger docs"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run Docker container"
	@echo "  mocks          - Generate mock implementations for testing"
	@echo "  dev-deps       - Install development dependencies"
	@echo "  migrate-create - Create a new migration file"
	@echo "  migrate-up     - Run migrations up"
	@echo "  migrate-down   - Run migrations down"
	@echo "  migrate-force  - Force migration version to fix dirty state"
	@echo "  help           - Show this help message"

.PHONY: build test clean run wire proto-install proto-clean proto-gen proto-swagger \
        lint fmt vet docker-build docker-run dev-deps test-coverage mocks help \
        migrate-create migrate-up migrate-down migrate-force