package display

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// openI2CBus opens an I2C bus device and resolves the device address, applying
// defaultAddr when the config leaves i2cAddr empty. Callers that keep the bus
// around should store the returned closer so Close can release it.
func openI2CBus(i2cBus, i2cAddr string, defaultAddr uint16) (i2c.BusCloser, uint16, error) {
	if _, err := host.Init(); err != nil {
		return nil, 0, fmt.Errorf("failed to initialize periph: %w", err)
	}

	bus, err := i2creg.Open(i2cBus)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open I2C bus %s: %w", i2cBus, err)
	}

	addr := defaultAddr
	if s := strings.TrimSpace(i2cAddr); s != "" {
		parsed, err := parseI2CAddr(s)
		if err != nil {
			bus.Close() // #nosec G104 -- best-effort cleanup on error path
			return nil, 0, err
		}
		addr = parsed
	}
	return bus, addr, nil
}

// Display is the interface for OLED display operations
type Display interface {
	// Init initializes the display hardware
	Init() error

	// Clear clears the image buffer without flushing to the display.
	// Call Show() afterwards to update the physical display.
	Clear() error

	// DrawText draws text at the specified position
	DrawText(x, y int, text string, size int) error

	// DrawLine draws a horizontal line
	DrawLine(x, y, width int) error

	// DrawPixel draws a single pixel
	DrawPixel(x, y int, on bool) error

	// DrawRect draws a rectangle
	DrawRect(x, y, width, height int, fill bool) error

	// DrawImage draws an image at the specified position
	DrawImage(x, y int, img image.Image) error

	// Show flushes the buffer to the display
	Show() error

	// Close closes the display
	Close() error

	// GetBounds returns the display dimensions
	GetBounds() image.Rectangle

	// GetBuffer returns a copy of the current display buffer (for testing)
	GetBuffer() []byte

	// SetBrightness sets the display brightness (0-255)
	SetBrightness(level uint8) error
}

// Font sizes
const (
	FontSmall  = 8
	FontMedium = 12
	FontLarge  = 16
)

// Color definitions
var (
	ColorOn  = color.White
	ColorOff = color.Black
)
