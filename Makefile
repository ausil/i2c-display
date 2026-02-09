.PHONY: build test clean install uninstall test-hardware

# Build configuration
BINARY_NAME=ssd1306d
BUILD_DIR=bin
INSTALL_DIR=/usr/local/bin
CONFIG_DIR=/etc/ssd1306-display
SYSTEMD_DIR=/etc/systemd/system

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ssd1306d/
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -cover ./...

# Run hardware tests (only on device with actual hardware)
test-hardware:
	@echo "Running hardware tests..."
	$(GOTEST) -v -tags=hardware ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Install the binary, config, and systemd service
install: build
	@echo "Installing $(BINARY_NAME)..."
	@sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@sudo mkdir -p $(CONFIG_DIR)
	@sudo cp configs/config.example.json $(CONFIG_DIR)/config.json
	@sudo cp systemd/ssd1306-display.service $(SYSTEMD_DIR)/
	@sudo systemctl daemon-reload
	@echo "Installation complete"
	@echo ""
	@echo "To enable and start the service:"
	@echo "  sudo systemctl enable ssd1306-display.service"
	@echo "  sudo systemctl start ssd1306-display.service"
	@echo ""
	@echo "To check status:"
	@echo "  sudo systemctl status ssd1306-display.service"

# Uninstall the binary, config, and systemd service
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo systemctl stop ssd1306-display.service 2>/dev/null || true
	@sudo systemctl disable ssd1306-display.service 2>/dev/null || true
	@sudo rm -f $(SYSTEMD_DIR)/ssd1306-display.service
	@sudo rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstall complete (config preserved in $(CONFIG_DIR))"

# Cross-compile for Raspberry Pi (32-bit ARM)
build-arm7:
	@echo "Building for ARMv7..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-arm7 ./cmd/ssd1306d/
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-arm7"

# Cross-compile for Raspberry Pi 4 / Rock 3C (64-bit ARM)
build-arm64:
	@echo "Building for ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-arm64 ./cmd/ssd1306d/
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-arm64"

# Build all architectures
build-all: build build-arm7 build-arm64

# Run with mock display (for testing without hardware)
run-mock: build
	@echo "Running with mock display..."
	$(BUILD_DIR)/$(BINARY_NAME) -mock -config configs/config.example.json

# Show help
help:
	@echo "SSD1306 Display Controller Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build the binary for current architecture"
	@echo "  test          - Run unit tests"
	@echo "  test-hardware - Run hardware tests (requires actual display)"
	@echo "  clean         - Remove build artifacts"
	@echo "  install       - Install binary, config, and systemd service"
	@echo "  uninstall     - Remove binary and systemd service"
	@echo "  build-arm7    - Cross-compile for Raspberry Pi (32-bit ARM)"
	@echo "  build-arm64   - Cross-compile for ARM64 (Pi 4, Rock 3C)"
	@echo "  build-all     - Build for all architectures"
	@echo "  run-mock      - Run with mock display (no hardware needed)"
	@echo "  help          - Show this help message"
