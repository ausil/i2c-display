# I2C Display Controller

A Go application for Single Board Computers that controls I2C OLED displays, showing system stats and network information with rotating pages.

**Works with any SBC that provides I2C devices:**
- **Raspberry Pi** (all models with I2C support)
- **Radxa** (Rock 3C, Rock 5B, Rock 4, etc.)
- **Orange Pi** (all models)
- **Pine64** (all models)
- **Banana Pi**, **Odroid**, and any other Linux SBC with I2C support

## Supported Displays

### Fully Working ✅
- **SSD1306** - 128x64, 128x32, or 96x16 monochrome OLED
  - Most common I2C OLED display
  - Full support via periph.io
  - Types: `ssd1306`, `ssd1306_128x64`, `ssd1306_128x32`, `ssd1306_96x16`

### Framework Ready (Drivers Needed) 🔧
- **SH1106** - 128x64 monochrome (similar to SSD1306)
- **SSD1327** - 128x128 4-bit grayscale OLED
- **SSD1331** - 96x64 16-bit color OLED

These types are recognized and dimensions auto-set, but return a helpful error message explaining the driver is not yet implemented.

See [DISPLAY_TYPES.md](DISPLAY_TYPES.md) for detailed information and how to add new display drivers.

## Features

- **System Monitoring**: Display disk usage, RAM usage, and CPU temperature
- **Network Information**: Show IP addresses for configured network interfaces
- **Rotating Pages**: Automatically cycle through information pages
- **Flexible Configuration**: JSON-based configuration with multiple search paths and hot reload
- **Systemd Integration**: Run as a system service with automatic start
- **Hardware Abstraction**: Mock display for testing without physical hardware
- **Comprehensive Testing**: Over 80% test coverage with CI/CD support
- **Structured Logging**: JSON and console logging with configurable levels
- **Prometheus Metrics**: Optional metrics endpoint for monitoring
- **Error Handling**: Automatic retry with exponential backoff for I2C errors
- **Health Monitoring**: Component health tracking and status reporting
- **Config Validation**: Validate configuration without running the service

## Requirements

- Go 1.19 or later (for building from source)
- Supported OLED display (see Supported Displays section)
- Any Linux-based SBC with I2C support
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

**Radxa (Rock 3C, Rock 5B, etc.):**
```bash
# I2C is usually enabled by default in modern images
# Verify with: ls /dev/i2c-*
```

**Orange Pi / Pine64 / Other SBCs:**
```bash
# Most modern Linux images have I2C enabled by default
# Check your board's documentation if I2C devices are not present
# Verify with: ls /dev/i2c-*
```

**Verify I2C is working:**
```bash
# Install i2c-tools if not already installed
# For Debian/Ubuntu/Raspberry Pi OS:
sudo apt-get install -y i2c-tools

# For Fedora/RHEL/CentOS:
sudo dnf install -y i2c-tools
# or: sudo yum install -y i2c-tools

# Detect I2C devices (replace '1' with your I2C bus number if different)
sudo i2cdetect -y 1
# You should see your display address (typically 0x3C or 0x3D)
```

**Finding your I2C bus:**
```bash
# List all I2C buses
ls /dev/i2c-*
# Common values: /dev/i2c-0, /dev/i2c-1, /dev/i2c-3, etc.
# Use the bus number in your config.json
```

## Installation

### For ARM/aarch64 Single Board Computers (Recommended)

**Build and install packages locally on your SBC** (Raspberry Pi, Radxa, Orange Pi, Pine64, etc.):

Building packages locally on ARM is fast and integrates better with your system's package manager.

<details>
<summary><strong>Debian/Ubuntu/Raspberry Pi OS (DEB Package)</strong></summary>

```bash
# Install build dependencies
sudo apt-get update
sudo apt-get install -y git golang debhelper devscripts dh-golang

# Clone and build
git clone https://github.com/ausil/i2c-display.git
cd i2c-display
make deb

# Install the package
sudo apt install ../i2c-display_*_arm64.deb  # or *_armhf.deb for 32-bit

# Enable and start
sudo systemctl enable --now i2c-display.service
```

The package automatically:
- Installs the binary to `/usr/bin/i2c-displayd`
- Creates config at `/etc/i2c-display/config.json`
- Installs and enables the systemd service
- Can be removed with: `sudo apt remove i2c-display`

</details>

<details>
<summary><strong>Fedora/RHEL on ARM (RPM Package)</strong></summary>

```bash
# Install build dependencies
sudo dnf install -y git golang rpm-build systemd-rpm-macros

# Clone and build
git clone https://github.com/ausil/i2c-display.git
cd i2c-display
make rpm

# Install the package
sudo dnf install rpm-build/RPMS/*/i2c-display-*.aarch64.rpm

# Enable and start
sudo systemctl enable --now i2c-display.service
```

The package automatically:
- Installs the binary to `/usr/bin/i2c-displayd`
- Creates config at `/etc/i2c-display/config.json`
- Installs and enables the systemd service
- Can be removed with: `sudo dnf remove i2c-display`

</details>

**Alternative: Pre-built Binaries**

If you prefer not to build packages, you can use pre-built binaries:

```bash
# Download for ARM64 (e.g., Raspberry Pi 4, Rock 5B)
wget https://github.com/ausil/i2c-display/releases/latest/download/i2c-displayd-linux-arm64
chmod +x i2c-displayd-linux-arm64
sudo mv i2c-displayd-linux-arm64 /usr/bin/i2c-displayd

# Download for ARMv7 (e.g., Raspberry Pi 2/3)
# wget https://github.com/ausil/i2c-display/releases/latest/download/i2c-displayd-linux-armv7
# chmod +x i2c-displayd-linux-armv7
# sudo mv i2c-displayd-linux-armv7 /usr/bin/i2c-displayd

# Create config directory and download config
sudo mkdir -p /etc/i2c-display
sudo wget -O /etc/i2c-display/config.json \
  https://raw.githubusercontent.com/ausil/i2c-display/main/configs/config.example.json

# Download and install systemd service
sudo wget -O /etc/systemd/system/i2c-display.service \
  https://raw.githubusercontent.com/ausil/i2c-display/main/systemd/i2c-display.service

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable --now i2c-display.service
```

**Note:** Using packages is recommended as they integrate with your system's package manager for easier updates and removal.

### For x86_64 Systems

**Pre-built packages available** from [GitHub Releases](https://github.com/ausil/i2c-display/releases/latest):

<details>
<summary><strong>RPM (Fedora, RHEL, Rocky Linux, AlmaLinux, CentOS)</strong></summary>

```bash
# Download the RPM
wget https://github.com/ausil/i2c-display/releases/latest/download/i2c-display-*.x86_64.rpm

# Install using dnf (Fedora 22+, RHEL 8+)
sudo dnf install ./i2c-display-*.x86_64.rpm
sudo systemctl enable --now i2c-display.service

# Or using yum (older RHEL/CentOS)
sudo yum localinstall i2c-display-*.x86_64.rpm
sudo systemctl enable --now i2c-display.service

# Or using rpm directly
sudo rpm -Uvh i2c-display-*.x86_64.rpm
sudo systemctl enable --now i2c-display.service
```

</details>

<details>
<summary><strong>DEB (Debian, Ubuntu, Linux Mint, Pop!_OS)</strong></summary>

```bash
# Download the DEB
wget https://github.com/ausil/i2c-display/releases/latest/download/i2c-display_*_amd64.deb

# Install using apt (recommended)
sudo apt install ./i2c-display_*_amd64.deb
sudo systemctl enable --now i2c-display.service

# Or using dpkg
sudo dpkg -i i2c-display_*_amd64.deb
sudo apt-get install -f  # Install any missing dependencies
sudo systemctl enable --now i2c-display.service
```

</details>

**Or use the pre-built binary:**

```bash
wget https://github.com/ausil/i2c-display/releases/latest/download/i2c-displayd-linux-amd64
chmod +x i2c-displayd-linux-amd64
sudo mv i2c-displayd-linux-amd64 /usr/bin/i2c-displayd
# Then follow manual setup steps from ARM binary installation above
```

### Building from Source

**Quick Install:**
```bash
git clone https://github.com/ausil/i2c-display.git
cd i2c-display
sudo ./scripts/install.sh
```

**Manual Build:**
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
    "output": "stdout",
    "json": false
  },
  "metrics": {
    "enabled": false,
    "address": ":9090"
  }
}
```

### Configuration Options

#### Display

- **`type`**: Display controller type (default: `ssd1306`)
  - `ssd1306` or `ssd1306_128x64` - Standard 128x64 OLED
  - `ssd1306_128x32` - Compact 128x32 OLED
  - `ssd1306_96x16` - Small 96x16 OLED
  - See [DISPLAY_TYPES.md](DISPLAY_TYPES.md) for all supported types

- **`i2c_bus`**: I2C bus device path (default: `/dev/i2c-1`)
  - Find your bus: `ls /dev/i2c-*`
  - Common values: `/dev/i2c-0`, `/dev/i2c-1`, `/dev/i2c-3`

- **`i2c_address`**: I2C device address in hexadecimal (default: `0x3C`)
  - Detect with: `sudo i2cdetect -y 1`
  - Common addresses: `0x3C` or `0x3D`

- **`rotation`**: Display rotation in 90° increments (default: `0`)
  - `0` - Normal orientation
  - `1` - Rotated 90° clockwise
  - `2` - Rotated 180° (upside down)
  - `3` - Rotated 270° clockwise (90° counter-clockwise)

- **`width`** / **`height`**: Display dimensions in pixels (optional)
  - **Automatically set** based on display type - no need to specify
  - Only needed for custom/unsupported displays

#### Pages

- **`rotation_interval`**: How often to rotate between pages
  - Format: Duration string (e.g., `"5s"`, `"30s"`, `"2m"`)
  - Default: `"5s"`

- **`refresh_interval`**: How often to update data on current page
  - Format: Duration string (e.g., `"1s"`, `"500ms"`)
  - Default: `"1s"`

#### System Info

- **`hostname_display`**: How to display the hostname
  - `"short"` - Only hostname (e.g., `raspberrypi`)
  - `"full"` - Full FQDN (e.g., `raspberrypi.local`)

- **`disk_path`**: Filesystem path to monitor (default: `"/"`)
  - Examples: `"/"`, `"/home"`, `"/mnt/data"`

- **`temperature_source`**: Path to CPU temperature sensor
  - **Raspberry Pi**: `/sys/class/thermal/thermal_zone0/temp`
  - **Radxa Rock 5B**: `/sys/class/thermal/thermal_zone0/temp`
  - **Orange Pi**: `/sys/class/thermal/thermal_zone0/temp` or `/sys/devices/virtual/thermal/thermal_zone0/temp`
  - **Pine64**: Check `ls /sys/class/thermal/thermal_zone*/temp`
  - Leave empty (`""`) to disable temperature display

- **`temperature_unit`**: Display unit for temperature
  - `"celsius"` - Display in °C
  - `"fahrenheit"` - Display in °F

**Finding your temperature sensor:**
```bash
# List all thermal zones
for zone in /sys/class/thermal/thermal_zone*/temp; do
  echo "$zone: $(cat $zone)"
done
```

#### Network

- **`auto_detect`**: Automatically find network interfaces (default: `true`)
  - Set to `false` to manually specify interfaces

- **`interface_filter.include`**: Patterns for interfaces to show
  - Supports wildcards: `["eth*", "wlan*", "usb*"]`
  - Default: `["eth0", "wlan0", "usb0"]`

- **`interface_filter.exclude`**: Patterns for interfaces to hide
  - Supports wildcards: `["lo", "docker*", "veth*"]`
  - Useful to hide virtual interfaces
  - Default: `["lo", "docker*", "veth*"]`

- **`show_ipv4`**: Display IPv4 addresses (default: `true`)

- **`show_ipv6`**: Display IPv6 addresses (default: `false`)

- **`max_interfaces_per_page`**: Maximum network interfaces per page (default: `3`)

**Example interface configurations:**

<details>
<summary>Show only Ethernet</summary>

```json
"network": {
  "auto_detect": true,
  "interface_filter": {
    "include": ["eth*", "en*"],
    "exclude": ["lo", "wlan*", "docker*", "veth*"]
  }
}
```
</details>

<details>
<summary>Show WiFi and Ethernet</summary>

```json
"network": {
  "auto_detect": true,
  "interface_filter": {
    "include": ["eth*", "wlan*", "en*", "wl*"],
    "exclude": ["lo", "docker*", "veth*"]
  }
}
```
</details>

#### Screen Saver (Optional)

Power saving feature to dim or blank the display after inactivity.

- **`enabled`**: Enable screen saver (default: `false`)

- **`mode`**: Screen saver behavior
  - `"dim"` - Reduce brightness
  - `"blank"` - Turn off display completely
  - `"off"` - No screen saver

- **`idle_timeout`**: Time before activating screen saver
  - Format: Duration string (e.g., `"5m"`, `"30m"`, `"1h"`)
  - Default: `"5m"`

- **`dim_brightness`**: Brightness level when dimmed (0-255)
  - `0` - Completely off
  - `50` - Half brightness
  - `255` - Full brightness
  - Default: `50`

- **`normal_brightness`**: Normal operating brightness (0-255)
  - Default: `255`

**Example:**
```json
"screensaver": {
  "enabled": true,
  "mode": "dim",
  "idle_timeout": "10m",
  "dim_brightness": 30,
  "normal_brightness": 255
}
```

#### Logging

- **`level`**: Log level verbosity
  - `"debug"` - Very verbose, includes all details
  - `"info"` - Normal operation information
  - `"warn"` - Warnings only
  - `"error"` - Errors only
  - Default: `"info"`

- **`output`**: Where to send logs
  - `"stdout"` - Standard output
  - `"stderr"` - Standard error
  - Default: `"stdout"`

- **`json`**: Log format
  - `true` - JSON format (good for log aggregation)
  - `false` - Human-readable console format
  - Default: `false`

#### Metrics (Optional)

Prometheus-compatible metrics endpoint for monitoring.

- **`enabled`**: Enable metrics endpoint (default: `false`)

- **`address`**: HTTP server address and port
  - Format: `"host:port"` or `":port"`
  - Examples: `":9090"`, `"127.0.0.1:9090"`, `"0.0.0.0:9090"`
  - Default: `":9090"`

When enabled, metrics are available at `http://address/metrics`

**Example metrics:**
- Display update count and errors
- I2C communication metrics
- Page rotation statistics
- System resource usage

### Platform-Specific Configuration Examples

<details>
<summary><strong>Raspberry Pi with SSD1306 128x64</strong></summary>

```json
{
  "display": {
    "type": "ssd1306_128x64",
    "i2c_bus": "/dev/i2c-1",
    "i2c_address": "0x3C",
    "rotation": 0
  },
  "system_info": {
    "temperature_source": "/sys/class/thermal/thermal_zone0/temp",
    "temperature_unit": "celsius"
  },
  "network": {
    "auto_detect": true,
    "interface_filter": {
      "include": ["eth0", "wlan0"],
      "exclude": ["lo", "docker*"]
    }
  }
}
```
</details>

<details>
<summary><strong>Radxa Rock 5B</strong></summary>

```json
{
  "display": {
    "type": "ssd1306_128x64",
    "i2c_bus": "/dev/i2c-7",
    "i2c_address": "0x3C",
    "rotation": 0
  },
  "system_info": {
    "temperature_source": "/sys/class/thermal/thermal_zone0/temp",
    "temperature_unit": "celsius"
  },
  "network": {
    "auto_detect": true,
    "interface_filter": {
      "include": ["eth*", "wlan*", "en*"],
      "exclude": ["lo", "docker*", "veth*"]
    }
  }
}
```
</details>

<details>
<summary><strong>Orange Pi with Smaller Display (128x32)</strong></summary>

```json
{
  "display": {
    "type": "ssd1306_128x32",
    "i2c_bus": "/dev/i2c-0",
    "i2c_address": "0x3C",
    "rotation": 0
  },
  "pages": {
    "rotation_interval": "10s",
    "refresh_interval": "2s"
  },
  "system_info": {
    "temperature_source": "/sys/devices/virtual/thermal/thermal_zone0/temp",
    "temperature_unit": "celsius"
  }
}
```
</details>

<details>
<summary><strong>Pine64 with Screen Saver</strong></summary>

```json
{
  "display": {
    "type": "ssd1306_128x64",
    "i2c_bus": "/dev/i2c-1",
    "i2c_address": "0x3C",
    "rotation": 0
  },
  "screensaver": {
    "enabled": true,
    "mode": "dim",
    "idle_timeout": "5m",
    "dim_brightness": 30,
    "normal_brightness": 255
  }
}
```
</details>

<details>
<summary><strong>Server/Headless Setup with Metrics</strong></summary>

```json
{
  "display": {
    "type": "ssd1306_128x64",
    "i2c_bus": "/dev/i2c-1",
    "i2c_address": "0x3C",
    "rotation": 0
  },
  "logging": {
    "level": "info",
    "output": "stdout",
    "json": true
  },
  "metrics": {
    "enabled": true,
    "address": ":9090"
  }
}
```

Access metrics: `curl http://localhost:9090/metrics`
</details>

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

# Validate configuration without running
./bin/i2c-displayd -validate-config -config /path/to/config.json

# Reload configuration (send SIGHUP to running process)
sudo systemctl reload i2c-display.service
# Or: sudo kill -HUP $(pidof i2c-displayd)
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

## Monitoring

### Prometheus Metrics

Enable metrics in your configuration:

```json
{
  "metrics": {
    "enabled": true,
    "address": "127.0.0.1:9090"
  }
}
```

Available metrics:
- `i2c_display_refresh_total` - Total display refreshes
- `i2c_display_refresh_errors_total` - Display errors by type
- `i2c_display_refresh_latency_seconds` - Refresh latency histogram
- `i2c_display_i2c_errors_total` - I2C communication errors
- `i2c_display_cpu_temperature_celsius` - Current CPU temperature
- `i2c_display_memory_used_percent` - Memory usage percentage
- `i2c_display_disk_used_percent` - Disk usage percentage
- `i2c_display_network_interfaces_count` - Number of network interfaces
- `i2c_display_current_page` - Current page number
- `i2c_display_page_rotation_total` - Total page rotations

Access metrics: `curl http://127.0.0.1:9090/metrics`

### Logging

Structured logging with contextual information:

```bash
# Console output (human-readable)
{"level":"info","time":"2026-02-09T12:00:00Z","message":"Display service running"}

# JSON output (for log aggregation)
{
  "level":"info",
  "display_type":"ssd1306",
  "bus":"/dev/i2c-1",
  "time":"2026-02-09T12:00:00Z",
  "message":"Initializing display hardware"
}
```

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
