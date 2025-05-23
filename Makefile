# Makefile for go-user-service

.PHONY: build test clean run

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