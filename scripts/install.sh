#!/bin/bash
set -e

echo "SSD1306 Display Service Installation"
echo "====================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
   echo "Please run as root (use sudo)"
   exit 1
fi

# Build the binary
echo "Building binary..."
make build

# Install
echo "Installing..."
make install

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "1. Edit /etc/ssd1306-display/config.json to configure your display"
echo "2. Enable the service: sudo systemctl enable ssd1306-display.service"
echo "3. Start the service: sudo systemctl start ssd1306-display.service"
echo "4. Check status: sudo systemctl status ssd1306-display.service"
echo ""
