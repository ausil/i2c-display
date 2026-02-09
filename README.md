# I2C Display Controller

A Go application for Single Board Computers (Raspberry Pi 3/4, Rock 3C) that controls I2C OLED displays via I2C, showing system stats and network information with rotating pages.

## Supported Displays

- **SSD1306** - 128x64, 128x32, or 96x16 monochrome OLED (default)
- Extensible architecture for additional I2C displays

See [DISPLAY_TYPES.md](DISPLAY_TYPES.md) for detailed display information and how to add new types.

## Features

- **System Monitoring**: Display disk usage, RAM usage, and CPU temperature
- **Network Information**: Show IP addresses for configured network interfaces
- **Rotating Pages**: Automatically cycle through information pages
- **Flexible Configuration**: JSON-based configuration with multiple search paths
- **Systemd Integration**: Run as a system service with automatic start
- **Hardware Abstraction**: Mock display for testing without physical hardware
- **Comprehensive Testing**: Over 80% test coverage with CI/CD support

## Requirements

- Go 1.19 or later
- SSD1306 128x64 OLED display connected via I2C
- Linux-based SBC (Raspberry Pi, Rock 3C, etc.)
- I2C enabled on the system

## Hardware Setup

### Wiring

Connect the SSD1306 display to your SBC:

| SSD1306 Pin | SBC Pin     | Description |
|-------------|-------------|-------------|
| VCC         | 3.3V        | Power       |
| GND         | GND         | Ground      |
| SCL         | I2C SCL     | I2C Clock   |
| SDA         | I2C SDA     | I2C Data    |

### Enable I2C

**Raspberry Pi:**
```bash
sudo raspi-config
# Select: Interface Options -> I2C -> Enable
sudo reboot
```

**Rock 3C:**
```bash
# I2C is usually enabled by default
# Verify with: ls /dev/i2c-*
```

Verify I2C is working:
```bash
sudo i2cdetect -y 1
# You should see your display address (typically 0x3C or 0x3D)
```

## Installation

### Quick Install

```bash
git clone https://github.com/ausil/i2c-display.git
cd i2c-display
sudo ./scripts/install.sh
```

### Manual Installation

```bash
# Build the binary
make build

# Install
sudo make install

# Enable and start the service
sudo systemctl enable i2c-display.service
sudo systemctl start i2c-display.service
```

## Configuration

The configuration file is searched in the following order:

1. Path specified with `-config` flag
2. `$I2C_DISPLAY_CONFIG_PATH` environment variable
3. `/etc/i2c-display/config.json` (system-wide)
4. `$HOME/.config/i2c-display/config.json` (user-specific)
5. `./config.json` (current directory)

### Example Configuration

See `configs/config.example.json` for a complete example:

```json
{
  "display": {
    "type": "ssd1306",
    "i2c_bus": "/dev/i2c-1",
    "i2c_address": "0x3C",
    "rotation": 0
  },
  "pages": {
    "rotation_interval": "5s",
    "refresh_interval": "1s"
  },
  "system_info": {
    "hostname_display": "short",
    "disk_path": "/",
    "temperature_source": "/sys/class/thermal/thermal_zone0/temp",
    "temperature_unit": "celsius"
  },
  "network": {
    "auto_detect": true,
    "interface_filter": {
      "include": ["eth0", "wlan0", "usb0"],
      "exclude": ["lo", "docker*", "veth*"]
    },
    "show_ipv4": true,
    "show_ipv6": false,
    "max_interfaces_per_page": 3
  },
  "logging": {
    "level": "info",
    "output": "stdout"
  }
}
```

### Configuration Options

#### Display

- `type`: Display controller type (default: `ssd1306`)
  - `ssd1306` or `ssd1306_128x64` - Standard 128x64
  - `ssd1306_128x32` - Compact 128x32 variant
  - `ssd1306_96x16` - Small 96x16 variant
- `i2c_bus`: I2C bus device (default: `/dev/i2c-1`)
- `i2c_address`: I2C address in hex (default: `0x3C`)
- `width`: Display width in pixels (optional - auto-detected from type)
- `height`: Display height in pixels (optional - auto-detected from type)
- `rotation`: Display rotation 0-3 (default: 0)

**Note**: Width and height are automatically set based on the display type. You only need to specify them if using custom dimensions.

#### Pages

- `rotation_interval`: Time between page changes (e.g., "5s", "10s")
- `refresh_interval`: Time between data updates (e.g., "1s", "2s")

#### System Info

- `hostname_display`: "short" or "full"
- `disk_path`: Path to monitor disk usage (default: "/")
- `temperature_source`: Path to CPU temperature file
- `temperature_unit`: "celsius" or "fahrenheit"

#### Network

- `auto_detect`: Automatically detect interfaces
- `interface_filter.include`: Interface patterns to include
- `interface_filter.exclude`: Interface patterns to exclude
- `show_ipv4`: Show IPv4 addresses
- `show_ipv6`: Show IPv6 addresses
- `max_interfaces_per_page`: Max interfaces per page

#### Logging

- `level`: "debug", "info", "warn", or "error"
- `output`: "stdout" or "stderr"

## Usage

### Run as Service

```bash
# Start service
sudo systemctl start i2c-display.service

# Stop service
sudo systemctl stop i2c-display.service

# Restart service
sudo systemctl restart i2c-display.service

# Check status
sudo systemctl status i2c-display.service

# View logs
sudo journalctl -u i2c-display.service -f
```

### Run Manually

```bash
# With default config search
./bin/i2c-displayd

# With specific config
./bin/i2c-displayd -config /path/to/config.json

# With mock display (for testing)
./bin/i2c-displayd -mock -config configs/config.example.json
```

## Development

### Building

```bash
# Build for current architecture
make build

# Build for Raspberry Pi (32-bit ARM)
make build-arm7

# Build for Raspberry Pi 4 / Rock 3C (64-bit ARM)
make build-arm64

# Build all architectures
make build-all
```

### Testing

```bash
# Run unit tests
make test

# Run with mock display (no hardware needed)
make run-mock

# Run hardware tests (requires actual display)
make test-hardware
```

### Project Structure

```
i2c-display/
├── cmd/
│   └── i2c-displayd/           # Main application
├── internal/
│   ├── config/             # Configuration management
│   ├── display/            # Display abstraction layer
│   ├── stats/              # System statistics collectors
│   ├── renderer/           # Page rendering
│   └── rotation/           # Page rotation manager
├── configs/                # Example configurations
├── systemd/                # Systemd service file
├── scripts/                # Installation scripts
├── testdata/               # Test fixtures
├── Makefile
└── README.md
```

## Display Layout

### Page 1: System Stats

```
┌──────────────────────────┐
│      hostname            │ (centered)
├──────────────────────────┤
│ Disk: 45.2% (12.5/27.6GB)│
│ RAM: 62.8% (2.5/4GB)     │
│ CPU: 45.2°C              │
└──────────────────────────┘
```

### Page 2+: Network Interfaces

```
┌──────────────────────────┐
│      hostname            │ (centered)
├──────────────────────────┤
│ eth0: 192.168.1.100      │
│ wlan0: 10.0.0.50         │
│ usb0: 172.16.0.1         │
│                 Page 2/3 │
└──────────────────────────┘
```

## Troubleshooting

### Display Not Working

1. Check I2C is enabled:
   ```bash
   ls /dev/i2c-*
   ```

2. Check I2C address:
   ```bash
   sudo i2cdetect -y 1
   ```

3. Verify permissions:
   ```bash
   sudo usermod -a -G i2c $USER
   # Log out and back in
   ```

4. Check service logs:
   ```bash
   sudo journalctl -u i2c-display.service -n 50
   ```

### Temperature Not Showing

Different SBCs have different temperature sensor paths:

- Raspberry Pi: `/sys/class/thermal/thermal_zone0/temp`
- Rock 3C: `/sys/class/thermal/thermal_zone0/temp` or `/sys/devices/virtual/thermal/thermal_zone0/temp`

Update `temperature_source` in your config accordingly.

### Network Interfaces Not Showing

Check your interface filter settings in the config. Use:
```bash
ip link show
```
to see available interfaces and adjust the `include` patterns.

## Uninstallation

```bash
sudo ./scripts/uninstall.sh
```

Or manually:
```bash
sudo systemctl stop i2c-display.service
sudo systemctl disable i2c-display.service
sudo make uninstall
```

## License

BSD 3-Clause License. See [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `make test`
5. Submit a pull request

## Acknowledgments

- Built with [periph.io](https://periph.io/) for hardware abstraction
- Uses [basicfont](https://pkg.go.dev/golang.org/x/image/font/basicfont) for text rendering

## Support

For issues, questions, or contributions, please visit:
https://github.com/ausil/i2c-display
