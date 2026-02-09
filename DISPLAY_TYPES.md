# Supported Display Types

This document describes the I2C OLED displays currently supported and how to add new ones.

## Currently Supported ✅

### SSD1306 Family (via periph.io)

All SSD1306 variants are fully supported through the periph.io library:

| Type | Resolution | Description | Status |
|------|------------|-------------|--------|
| `ssd1306` | 128x64 | Default, most common variant | ✅ Working |
| `ssd1306_128x64` | 128x64 | Explicit 128x64 variant | ✅ Working |
| `ssd1306_128x32` | 128x32 | Smaller variant, common in compact displays | ✅ Working |
| `ssd1306_96x16` | 96x16 | Very small variant | ✅ Working |

## Framework Ready (Drivers Needed) 🔧

These displays are recognized by the configuration system but need driver implementations:

| Type | Resolution | Color | Driver Status |
|------|------------|-------|---------------|
| `sh1106` | 128x64 | Monochrome | Third-party available (SPI only) |
| `sh1106_128x64` | 128x64 | Monochrome | Third-party available (SPI only) |
| `ssd1327` | 128x128 | 4-bit grayscale | No Go I2C driver found |
| `ssd1327_96x96` | 96x96 | 4-bit grayscale | No Go I2C driver found |
| `ssd1331` | 96x64 | 16-bit color | No Go I2C driver found |

**Note:** The configuration will accept these types and auto-set dimensions, but will return an error message explaining the driver is not implemented.

### Configuration

Set the display type in your config file. Width and height are automatically determined:

```json
{
  "display": {
    "type": "ssd1306",
    "i2c_bus": "/dev/i2c-1",
    "i2c_address": "0x3C",
    "rotation": 0
  }
}
```

For a 128x32 display:

```json
{
  "display": {
    "type": "ssd1306_128x32",
    "i2c_bus": "/dev/i2c-1",
    "i2c_address": "0x3C",
    "rotation": 0
  }
}
```

The dimensions (128x32) are automatically set based on the display type.

### Example Configurations

- **Standard 128x64**: `configs/config.example.json`
- **Compact 128x32**: `configs/config.ssd1306_128x32.json`

## Display Type Detection

The application automatically selects the correct driver based on the `type` field. All `ssd1306*` types use the same periph.io driver with different dimensions.

## Driver Availability Research

Based on searches of available Go libraries:

### SH1106
- **GitHub:** [danielgatis/go-sh1106](https://github.com/danielgatis/go-sh1106) - **SPI only**, not I2C
- **Alternative:** [sandbankdisperser/go-i2c-oled](https://pkg.go.dev/github.com/sandbankdisperser/go-i2c-oled/sh1106) - May support I2C
- **Status:** Possible to add with third-party library
- **Compatibility:** Very similar to SSD1306, just different memory mapping

### SSD1327 (Grayscale)
- **Status:** No mature Go I2C drivers found
- **Python:** Many Python libraries exist (Adafruit, Luma.OLED)
- **Would need:** Port from Python or write from datasheet

### SSD1331/SSD1351 (Color)
- **Status:** No Go I2C drivers found
- **Note:** Most color OLEDs use SPI, not I2C
- **Would need:** Implement from datasheet

## Adding New Display Types

To add support for additional I2C OLED displays:

### 1. Check periph.io Support

First, check if the display is supported by periph.io:
- Visit: https://pkg.go.dev/periph.io/x/devices/v3
- Look for your display controller (e.g., sh1106, ssd1327)

### 2. Create Display Implementation

If periph.io supports it, create a new file like `internal/display/sh1106.go`:

```go
package display

import (
    "periph.io/x/devices/v3/sh1106"
    // ... other imports
)

type SH1106Display struct {
    dev *sh1106.Dev
    // ... fields
}

func NewSH1106Display(i2cBus string, width, height, rotation int) (*SH1106Display, error) {
    // Implementation similar to ssd1306.go
}

// Implement all Display interface methods
```

### 3. Update Factory

Add the new type to `internal/display/factory.go`:

```go
func NewDisplay(cfg *config.DisplayConfig) (Display, error) {
    displayType := strings.ToLower(cfg.Type)

    if strings.HasPrefix(displayType, "ssd1306") {
        return NewSSD1306Display(...)
    }

    // Add new display type
    if strings.HasPrefix(displayType, "sh1106") {
        return NewSH1106Display(...)
    }

    return nil, fmt.Errorf("unsupported display type: %s", cfg.Type)
}
```

### 4. Update Configuration Validation

Add the new type to `internal/config/config.go`:

```go
validTypes := map[string]bool{
    "ssd1306":        true,
    "ssd1306_128x32": true,
    "ssd1306_128x64": true,
    "ssd1306_96x16":  true,
    "sh1106":         true,  // NEW
    "sh1106_128x64":  true,  // NEW
}
```

### 5. Add Example Configuration

Create `configs/config.sh1106.json` with appropriate settings.

### 6. Update Documentation

Update README.md with the new display type.

## Third-Party Display Drivers

If periph.io doesn't support your display, you can use third-party libraries:

### SH1106 (Third-Party)

The [danielgatis/go-sh1106](https://github.com/danielgatis/go-sh1106) library provides SH1106 support using periph.io's I2C interfaces.

To add it:
```bash
go get github.com/danielgatis/go-sh1106
```

Then create `internal/display/sh1106_thirdparty.go` using this library.

## Common I2C Addresses

| Display | Common Address |
|---------|----------------|
| SSD1306 | 0x3C or 0x3D |
| SH1106  | 0x3C or 0x3D |
| SSD1327 | 0x3C or 0x3D |

Use `i2cdetect -y 1` on your SBC to find the actual address.

## Display Differences

### Resolution

Different displays support different resolutions. Adjust the renderer based on available space:
- 128x64: Can show ~8 lines of text
- 128x32: Can show ~4 lines of text
- 96x16: Very limited, ~2 lines

### Color Depth

- **SSD1306**: Monochrome (1-bit)
- **SSD1327**: 4-bit grayscale (16 levels)
- **SSD1331**: 16-bit color (RGB565)

The current implementation is optimized for monochrome displays. Grayscale/color support would require renderer updates.

### Communication

All supported displays use I2C. SPI displays would require significant changes to the display layer.

## Resources

- [periph.io Devices](https://pkg.go.dev/periph.io/x/devices/v3)
- [periph.io SSD1306 Docs](https://periph.io/device/ssd1306/)
- [I2C Display Guide](https://learn.adafruit.com/adafruit-pioled-128x32-mini-oled-for-raspberry-pi)

## Sources

- [periph.io devices v3](https://pkg.go.dev/periph.io/x/devices/v3)
- [SSD1306 driver](https://pkg.go.dev/periph.io/x/devices/v3/ssd1306)
- [danielgatis/go-sh1106 (third-party)](https://github.com/danielgatis/go-sh1106)
