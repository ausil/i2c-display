.PHONY: build test clean install uninstall test-hardware dist rpm srpm deb deb-src lint fmt

# Version - prefer git tag if available (for releases), otherwise use VERSION file
GIT_TAG_VERSION=$(shell git describe --tags --exact-match 2>/dev/null | sed 's/^v//')
VERSION=$(or $(GIT_TAG_VERSION),$(shell cat VERSION))
PROJECT_NAME=i2c-display

# Build configuration
BINARY_NAME=i2c-displayd
BUILD_DIR=bin
DIST_DIR=dist
INSTALL_DIR=/usr/bin
CONFIG_DIR=/etc/i2c-display
SYSTEMD_DIR=/usr/lib/systemd/system

# RPM configuration
RPM_TOPDIR=$(shell pwd)/rpm-build
TARBALL=$(PROJECT_NAME)-$(VERSION).tar.gz

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/i2c-displayd/
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run linters
lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin" && exit 1)
	golangci-lint run --timeout=5m ./...

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w -local github.com/ausil/i2c-display .

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
	rm -rf $(BUILD_DIR) $(DIST_DIR) $(RPM_TOPDIR)
	rm -f ../i2c-display_*.deb ../i2c-display_*.dsc ../i2c-display_*.tar.xz ../i2c-display_*.changes ../i2c-display_*.buildinfo
	rm -rf debian/.debhelper debian/i2c-display debian/*.debhelper* debian/*.substvars debian/debhelper-build-stamp debian/files
	@echo "Clean complete"

# Install the binary, config, and systemd service
# Run with: sudo make install
install: build
	@echo "Installing $(BINARY_NAME)..."
	install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(DESTDIR)$(INSTALL_DIR)/$(BINARY_NAME)
	mkdir -p $(DESTDIR)$(CONFIG_DIR)
	cp configs/config.example.json $(DESTDIR)$(CONFIG_DIR)/config.json
	cp systemd/i2c-display.service $(DESTDIR)$(SYSTEMD_DIR)/
	mkdir -p $(DESTDIR)/usr/share/man/man1
	install -m 644 man/i2c-displayd.1 $(DESTDIR)/usr/share/man/man1/
	@echo "Installation complete"
	@echo ""
	@echo "To enable and start the service:"
	@echo "  systemctl daemon-reload"
	@echo "  systemctl enable i2c-display.service"
	@echo "  systemctl start i2c-display.service"

# Uninstall the binary, config, and systemd service
# Run with: sudo make uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	systemctl stop i2c-display.service 2>/dev/null || true
	systemctl disable i2c-display.service 2>/dev/null || true
	rm -f $(DESTDIR)$(SYSTEMD_DIR)/i2c-display.service
	rm -f $(DESTDIR)$(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstall complete (config preserved in $(CONFIG_DIR))"

# Cross-compile for Raspberry Pi (32-bit ARM)
build-arm7:
	@echo "Building for ARMv7..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-arm7 ./cmd/i2c-displayd/
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-arm7"

# Cross-compile for Raspberry Pi 4 / Rock 3C (64-bit ARM)
build-arm64:
	@echo "Building for ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-arm64 ./cmd/i2c-displayd/
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-arm64"

# Build all architectures
build-all: build build-arm7 build-arm64

# Run with mock display (for testing without hardware)
run-mock: build
	@echo "Running with mock display..."
	$(BUILD_DIR)/$(BINARY_NAME) -mock -config configs/config.example.json

# Create release tarball
dist:
	@echo "Creating release tarball v$(VERSION)..."
	@mkdir -p $(DIST_DIR)
	@rm -rf $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)
	@mkdir -p $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)
	@go mod vendor
	@cp -r cmd internal configs systemd scripts testdata vendor rpm debian man $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)/
	@cp go.mod go.sum Makefile VERSION LICENSE $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)/
	@cp README.md BUILDING.md CONTRIBUTING.md DISPLAY_TYPES.md LICENSES.md SECURITY.md $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)/
	@tar -czf $(DIST_DIR)/$(TARBALL) -C $(DIST_DIR) $(PROJECT_NAME)-$(VERSION)
	@rm -rf $(DIST_DIR)/$(PROJECT_NAME)-$(VERSION)
	@echo "Release tarball created: $(DIST_DIR)/$(TARBALL)"
	@ls -lh $(DIST_DIR)/$(TARBALL)

# Build source RPM
srpm: dist
	@echo "Building source RPM..."
	@mkdir -p $(RPM_TOPDIR)/BUILD
	@mkdir -p $(RPM_TOPDIR)/RPMS
	@mkdir -p $(RPM_TOPDIR)/SOURCES
	@mkdir -p $(RPM_TOPDIR)/SPECS
	@mkdir -p $(RPM_TOPDIR)/SRPMS
	@cp $(DIST_DIR)/$(TARBALL) $(RPM_TOPDIR)/SOURCES/
	@cp rpm/$(PROJECT_NAME).spec $(RPM_TOPDIR)/SPECS/
	@rpmbuild --define "_topdir $(RPM_TOPDIR)" --define "_unitdir /usr/lib/systemd/system" --nodeps -bs $(RPM_TOPDIR)/SPECS/$(PROJECT_NAME).spec
	@echo "Source RPM created:"
	@ls -lh $(RPM_TOPDIR)/SRPMS/*.src.rpm

# Build binary RPM
rpm: dist
	@echo "Building binary RPM..."
	@mkdir -p $(RPM_TOPDIR)/BUILD
	@mkdir -p $(RPM_TOPDIR)/RPMS
	@mkdir -p $(RPM_TOPDIR)/SOURCES
	@mkdir -p $(RPM_TOPDIR)/SPECS
	@mkdir -p $(RPM_TOPDIR)/SRPMS
	@cp $(DIST_DIR)/$(TARBALL) $(RPM_TOPDIR)/SOURCES/
	@cp rpm/$(PROJECT_NAME).spec $(RPM_TOPDIR)/SPECS/
	@rpmbuild --define "_topdir $(RPM_TOPDIR)" --define "_unitdir /usr/lib/systemd/system" --nodeps -ba $(RPM_TOPDIR)/SPECS/$(PROJECT_NAME).spec
	@echo "RPM packages created:"
	@ls -lh $(RPM_TOPDIR)/RPMS/*/*.rpm
	@ls -lh $(RPM_TOPDIR)/SRPMS/*.src.rpm

# Install RPM (requires rpm to be built first)
# Run with: sudo make install-rpm
install-rpm:
	@echo "Installing RPM..."
	rpm -Uvh $(RPM_TOPDIR)/RPMS/*/*.rpm

# Build Debian source package
deb-src:
	@echo "Building Debian source package..."
	@dpkg-buildpackage -d -S -us -uc
	@echo "Debian source package created:"
	@ls -lh ../$(PROJECT_NAME)_*.dsc

# Build Debian binary package
deb:
	@echo "Building Debian binary package..."
	@dpkg-buildpackage -d -b -us -uc
	@echo "Debian package created:"
	@ls -lh ../$(PROJECT_NAME)_*.deb

# Install DEB (requires deb to be built first)
# Run with: sudo make install-deb
install-deb:
	@echo "Installing DEB..."
	dpkg -i ../$(PROJECT_NAME)_*.deb
	apt-get install -f -y

# Show version
version:
	@echo "$(VERSION)"

# Show help
help:
	@echo "I2C Display Controller Makefile"
	@echo ""
	@echo "Version: $(VERSION)"
	@echo ""
	@echo "Build targets:"
	@echo "  build         - Build the binary for current architecture"
	@echo "  build-arm7    - Cross-compile for Raspberry Pi (32-bit ARM)"
	@echo "  build-arm64   - Cross-compile for ARM64 (Pi 4, Rock 3C)"
	@echo "  build-all     - Build for all architectures"
	@echo ""
	@echo "Test targets:"
	@echo "  test          - Run unit tests"
	@echo "  test-hardware - Run hardware tests (requires actual display)"
	@echo "  run-mock      - Run with mock display (no hardware needed)"
	@echo "  lint          - Run golangci-lint"
	@echo "  fmt           - Format code with gofmt and goimports"
	@echo ""
	@echo "Release targets:"
	@echo "  dist          - Create release tarball"
	@echo "  srpm          - Build source RPM"
	@echo "  rpm           - Build binary and source RPM"
	@echo "  deb-src       - Build Debian source package"
	@echo "  deb           - Build Debian binary package"
	@echo ""
	@echo "Installation targets:"
	@echo "  install       - Install from source (binary, config, systemd service)"
	@echo "  install-rpm   - Install from RPM package"
	@echo "  install-deb   - Install from DEB package"
	@echo "  uninstall     - Remove binary and systemd service"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean         - Remove build artifacts"
	@echo "  version       - Show current version"
	@echo "  help          - Show this help message"
