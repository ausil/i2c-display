#!/bin/bash
set -e

echo "I2C Display Service Installation"
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
echo "1. Edit /etc/i2c-display/config.json to configure your display"
echo "2. Enable the service: sudo systemctl enable i2c-display.service"
echo "3. Start the service: sudo systemctl start i2c-display.service"
echo "4. Check status: sudo systemctl status i2c-display.service"
echo ""
