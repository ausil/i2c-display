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

### SH1106 — I2C monochrome OLED (native driver)

| Type | Resolution | Description | Status |
|------|------------|-------------|--------|
| `sh1106` | 128x64 | Default, the controller on most 1.3" I2C OLED modules | ✅ Working |
| `sh1106_128x64` | 128x64 | Explicit 128x64 variant | ✅ Working |

The SH1106 looks like an SSD1306 but is not drop-in compatible, so it has its own
driver (`internal/display/sh1106.go`) rather than reusing periph's SSD1306 one:

- It only supports page addressing; the SSD1306 horizontal/vertical addressing mode
  and column/page window commands do not exist.
- Its GDDRAM is 132 columns wide for a 128-pixel panel, so frames are written with a
  column offset of 2.
- It needs `0xAD 0x8B` to enable its internal DC-DC converter; the SSD1306 charge-pump
  command is ignored, which leaves the panel completely dark.

`rotation` supports `0` and `2` (180°), applied with the mirror commands. If the image
comes up left/right and top/bottom flipped, use `"rotation": 2`.

**Wiring:** VCC, GND, SCL, SDA to the I2C bus on your SBC.

The bus number depends on the board — `/dev/i2c-1` on Raspberry Pi, `/dev/i2c-8` on
Radxa Rock 5C (40-pin header pins 3 = SDA, 5 = SCL) and Rock 5A, `/dev/i2c-3` on
Rock 3C, `/dev/i2c-2` on Orange Pi 3B. Find yours with `ls /dev/i2c-*` and confirm the
panel answers with `sudo i2cdetect -y <bus>` (expect `0x3C` or `0x3D`). Tested platform
profiles live in [`configs/platforms/`](configs/platforms/).

**Example config:**
```json
{
  "display": {
    "type": "sh1106",
    "i2c_bus": "/dev/i2c-1",
    "i2c_address": "0x3C",
    "rotation": 0
  }
}
```

See `configs/config.sh1106.json` for a complete example.

### ST7735 Family — SPI colour TFT (native driver, no extra dependencies)

| Type | Resolution | Module | Col offset | Row offset |
|------|------------|--------|-----------|-----------|
| `st7735` | 128x160 | 1.8" (default) | 0 | 0 |
| `st7735_128x160` | 128x160 | 1.8" | 0 | 0 |
| `st7735_128x128` | 128x128 | 1.44" red tab | 2 | 3 |
| `st7735_160x80` | 160x80 | 0.96" (e.g. Waveshare) | 1 | 26 |

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

### UCTRONICS Family — I2C colour TFT (via onboard MCU bridge)

| Type | Resolution | Description | Status |
|------|------------|-------------|--------|
| `uctronics_colour` | 160x80 | Pi Rack Pro 0.96" colour TFT (SKU_RM0004) | ✅ Working |

The UCTRONICS Pi Rack Pro has an onboard MCU that bridges I2C to the ST7735 display internally. The host communicates with the MCU at I2C address `0x18` — no SPI, DC, or RST pins are needed.

**Example config:**
```json
{
  "display": {
    "type": "uctronics_colour",
    "i2c_bus": "/dev/i2c-1",
    "i2c_address": "0x18"
  }
}
```

See `configs/config.uctronics_colour.json` for a complete example.

---

## Framework Ready (Drivers Needed) 🔧

These displays are recognized by the configuration system (dimensions auto-set) but need driver implementations. They return a descriptive error until a driver is added.

| Type | Resolution | Interface | Color | Driver Status |
|------|------------|-----------|-------|---------------|
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
| UCTRONICS (colour) | `0x18` |

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
| UCTRONICS (colour) | I2C (MCU bridge) | 160x80 | Color | 16 |

---

## Resources

- [periph.io Devices](https://pkg.go.dev/periph.io/x/devices/v3)
- [periph.io SSD1306 Docs](https://periph.io/device/ssd1306/)
- [periph.io SPI Docs](https://pkg.go.dev/periph.io/x/conn/v3/spi)
- [ST7735 datasheet](https://www.displayfuture.com/Display/datasheet/controller/ST7735.pdf)
- [SH1106 datasheet](https://www.displayfuture.com/Display/datasheet/controller/SH1106.pdf)
