# Makefile for go-user-service

.PHONY: build test clean run wire

# Define variables
SERVICE_NAME = go-user-service
BUILD_DIR = ./bin
CMD_DIR = ./cmd/app

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
	cd $(CMD_DIR) && wire
	@echo "Wire code generation complete."

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