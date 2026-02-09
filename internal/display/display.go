package display

import (
	"image"
	"image/color"
)

// Display is the interface for OLED display operations
type Display interface {
	// Init initializes the display hardware
	Init() error

	// Clear clears the entire display
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
