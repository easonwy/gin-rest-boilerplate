# Makefile for go-user-service

.PHONY: build test clean run wire proto proto-clean proto-gen proto-swagger

# Define variables
SERVICE_NAME = go-user-service
BUILD_DIR = ./bin
CMD_DIR = ./cmd/server

# Build the service
build: 
	@echo "Building $(SERVICE_NAME)..."
	go build -o $(BUILD_DIR)/$(SERVICE_NAME) $(CMD_DIR)
	@echo "Build complete."

# Run tests
test:
	@echo "Running tests..."
	go test ./...
	@echo "Tests complete."

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete."

# Run the service
run: build
	@echo "Running $(SERVICE_NAME)..."
	$(BUILD_DIR)/$(SERVICE_NAME)

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
	rm -rf $(PROTO_OUT)/*
	rm -rf $(SWAGGER_OUT)/*

# Generate protobuf code and swagger docs
proto-gen: proto-clean
	cd $(PROTO_DIR) && buf generate

# Generate swagger docs
proto-swagger: proto-gen
	@echo "Swagger documentation generated at $(SWAGGER_OUT)/user_service.swagger.json"

# --- Database Migration --- 

MIGRATE_CLI = migrate
MIGRATE_DIR = ./migrations
DB_URL = $(DATABASE_URL) # Use DATABASE_URL environment variable

.PHONY: migrate-create migrate-up migrate-down

migrate-create:
	@echo "Creating migration file..."
	$(MIGRATE_CLI) create -ext sql -dir $(MIGRATE_DIR) $(name)

migrate-up:
	@echo "Running migrations up..."
	$(MIGRATE_CLI) -database $(DB_URL) -path $(MIGRATE_DIR) up

migrate-down:
	@echo "Running migrations down..."
	$(MIGRATE_CLI) -database $(DB_URL) -path $(MIGRATE_DIR) down