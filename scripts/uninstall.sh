#!/bin/bash
set -e

echo "I2C Display Service Uninstallation"
echo "======================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
   echo "Please run as root (use sudo)"
   exit 1
fi

# Uninstall
echo "Uninstalling..."
make uninstall

echo ""
echo "Uninstallation complete!"
echo ""
echo "Configuration files are preserved in /etc/i2c-display/"
echo "To remove them manually: sudo rm -rf /etc/i2c-display/"
echo ""
