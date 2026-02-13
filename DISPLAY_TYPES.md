# Supported Display Types

This document describes all displays currently supported and how to add new ones.

## Currently Supported ✅

### SSD1306 Family — I2C monochrome OLED (via periph.io)

| Type | Resolution | Description | Status |
|------|------------|-------------|--------|
| `ssd1306` | 128x64 | Default, most common variant | ✅ Working |
| `ssd1306_128x64` | 128x64 | Explicit 128x64 variant | ✅ Working |
| `ssd1306_128x32` | 128x32 | Smaller variant | ✅ Working |
| `ssd1306_96x16` | 96x16 | Very small variant | ✅ Working |

**Wiring:** VCC, GND, SCL, SDA to the I2C bus on your SBC.

**Example config:**
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

### ST7735 Family — SPI colour TFT (native driver, no extra dependencies)

| Type | Resolution | Module | Col offset | Row offset |
|------|------------|--------|-----------|-----------|
| `st7735` | 128x160 | 1.8" (default) | 0 | 0 |
| `st7735_128x160` | 128x160 | 1.8" | 0 | 0 |
| `st7735_128x128` | 128x128 | 1.44" red tab | 2 | 3 |
| `st7735_160x80` | 160x80 | 0.96" (e.g. Waveshare) | 0 | 24 |

Colour: white-on-black (RGB565), consistent with OLED rendering.

**Wiring:**

| ST7735 Pin | SBC Pin               | Description              |
|------------|-----------------------|--------------------------|
| VCC        | 3.3V                  | Power                    |
| GND        | GND                   | Ground                   |
| SCL/SCK    | SPI SCLK              | SPI Clock                |
| SDA/MOSI   | SPI MOSI              | SPI Data                 |
| CS         | SPI CS0 (CE0)         | Chip Select              |
| DC/RS      | Any GPIO (e.g. GPIO24) | Data/Command select     |
| RST        | Any GPIO (e.g. GPIO25) | Reset (optional)        |

**Example config:**
```json
{
  "display": {
    "type": "st7735_160x80",
    "spi_bus": "SPI0.0",
    "dc_pin": "GPIO24",
    "rst_pin": "GPIO25",
    "rotation": 0
  }
}
```

See `configs/config.st7735_160x80.json` and `configs/config.st7735_128x128.json` for complete examples.

---

## Framework Ready (Drivers Needed) 🔧

These displays are recognized by the configuration system (dimensions auto-set) but need driver implementations. They return a descriptive error until a driver is added.

| Type | Resolution | Interface | Color | Driver Status |
|------|------------|-----------|-------|---------------|
| `sh1106` / `sh1106_128x64` | 128x64 | I2C | Monochrome | Third-party available |
| `ssd1327` / `ssd1327_128x128` | 128x128 | I2C | 4-bit grayscale | No Go driver found |
| `ssd1327_96x96` | 96x96 | I2C | 4-bit grayscale | No Go driver found |
| `ssd1331` / `ssd1331_96x64` | 96x64 | SPI | 16-bit color | No Go driver found |

---

## Adding New Display Types

### 1. Add to display specs

Add an entry to the `specs` map in `internal/config/display_specs.go`:

```go
"mynewdisplay": {Width: 128, Height: 64},
```

### 2. Update connection type helpers (if new bus type)

If your display uses a bus type not yet covered, update `IsI2C()` or `IsSPI()` in
`internal/config/config.go`, or add a new helper (e.g. `IsCustomBus()`).

### 3. Create the driver

Create `internal/display/mynewdisplay.go` implementing all methods of the `Display` interface:

```go
type MyNewDisplay struct { ... }

func NewMyNewDisplay(...) (*MyNewDisplay, error) { ... }

// Init, Clear, DrawText, DrawLine, DrawPixel, DrawRect,
// DrawImage, Show, Close, GetBounds, GetBuffer, SetBrightness
```

See `internal/display/ssd1306.go` (I2C) or `internal/display/st7735.go` (SPI) as reference implementations.

### 4. Wire into the factory

Add a case in `internal/display/factory.go`:

```go
if strings.HasPrefix(displayType, "mynewdisplay") {
    return NewMyNewDisplay(cfg.SPIBus, cfg.DCPin, cfg.Width, cfg.Height, cfg.Rotation)
}
```

### 5. Update validation

If the new display requires new config fields, add validation to
`validateDisplay()` in `internal/config/config.go`.

### 6. Add tests and an example config

- Add the type to `internal/config/display_specs_test.go`
- Create `configs/config.mynewdisplay.json`

---

## Common I2C Addresses

| Display | Common Address |
|---------|----------------|
| SSD1306 | `0x3C` or `0x3D` |
| SH1106  | `0x3C` or `0x3D` |
| SSD1327 | `0x3C` or `0x3D` |

Use `sudo i2cdetect -y 1` on your SBC to find the actual address.

SPI displays (ST7735) do not use I2C addresses — use `spi_bus`, `dc_pin`, and optionally `rst_pin` instead.

---

## Display Comparison

| Display | Interface | Resolution | Color | Bits/pixel |
|---------|-----------|------------|-------|-----------|
| SSD1306 | I2C | 128x64 max | Monochrome | 1 |
| SH1106  | I2C | 128x64 | Monochrome | 1 |
| SSD1327 | I2C | 128x128 | Grayscale | 4 |
| SSD1331 | SPI | 96x64 | Color | 16 |
| ST7735  | SPI | up to 128x160 | Color | 16 |

---

## Resources

- [periph.io Devices](https://pkg.go.dev/periph.io/x/devices/v3)
- [periph.io SSD1306 Docs](https://periph.io/device/ssd1306/)
- [periph.io SPI Docs](https://pkg.go.dev/periph.io/x/conn/v3/spi)
- [ST7735 datasheet](https://www.displayfuture.com/Display/datasheet/controller/ST7735.pdf)
- [danielgatis/go-sh1106 (third-party SH1106)](https://github.com/danielgatis/go-sh1106)
