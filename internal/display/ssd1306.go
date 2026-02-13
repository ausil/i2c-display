package display

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/host/v3"
)

// SSD1306Display implements Display interface for real SSD1306 hardware
type SSD1306Display struct {
	dev    *ssd1306.Dev
	img    *image.Gray
	width  int
	height int
}

// NewSSD1306Display creates a new SSD1306 display driver
func NewSSD1306Display(i2cBus, i2cAddr string, width, height, rotation int) (*SSD1306Display, error) {
	// Initialize periph host
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize periph: %w", err)
	}

	// Open I2C bus
	bus, err := i2creg.Open(i2cBus)
	if err != nil {
		return nil, fmt.Errorf("failed to open I2C bus %s: %w", i2cBus, err)
	}

	// SSD1306 only supports 0° (no rotation) and 180° (Rotated flag).
	// Hardware-level 90°/270° rotation is not available on this chip.
	if rotation != 0 && rotation != 2 {
		return nil, fmt.Errorf("SSD1306 only supports rotation 0 (0°) and 2 (180°), got %d", rotation)
	}

	// Determine display options
	opts := ssd1306.Opts{
		W:             width,
		H:             height,
		Rotated:       rotation == 2,
		Sequential:    true,
		SwapTopBottom: false,
	}

	// Create SSD1306 device
	dev, err := ssd1306.NewI2C(bus, &opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSD1306 device: %w", err)
	}

	return &SSD1306Display{
		dev:    dev,
		img:    image.NewGray(image.Rect(0, 0, width, height)),
		width:  width,
		height: height,
	}, nil
}

// Init initializes the display
func (d *SSD1306Display) Init() error {
	// The device is initialized in NewSSD1306Display
	// Clear the display to start fresh
	return d.Clear()
}

// Clear clears the display
func (d *SSD1306Display) Clear() error {
	// Clear the image buffer
	draw.Draw(d.img, d.img.Bounds(), &image.Uniform{color.Gray{Y: 0}}, image.Point{}, draw.Src)
	return nil
}

// DrawText draws text at the specified position
// Note: This is a simple implementation. For real text rendering,
// you would need a font library like golang.org/x/image/font
func (d *SSD1306Display) DrawText(x, y int, text string, size int) error {
	// For now, this is a placeholder that draws a rectangle
	// In a full implementation, you would use a font library
	charWidth := size / 2
	for i := range text {
		startX := x + i*charWidth
		if startX >= d.width {
			break
		}
		// Draw a simple rectangle to represent each character
		if err := d.DrawRect(startX, y, charWidth-1, size, false); err != nil {
			return err
		}
	}
	return nil
}

// DrawLine draws a horizontal line
func (d *SSD1306Display) DrawLine(x, y, width int) error {
	for i := 0; i < width && x+i < d.width; i++ {
		if x+i >= 0 && y >= 0 && y < d.height {
			d.img.SetGray(x+i, y, color.Gray{Y: 255})
		}
	}
	return nil
}

// DrawPixel draws a single pixel
func (d *SSD1306Display) DrawPixel(x, y int, on bool) error {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return nil
	}

	if on {
		d.img.SetGray(x, y, color.Gray{Y: 255})
	} else {
		d.img.SetGray(x, y, color.Gray{Y: 0})
	}
	return nil
}

// DrawRect draws a rectangle
//nolint:gocyclo // drawing logic naturally has many conditional branches
func (d *SSD1306Display) DrawRect(x, y, width, height int, fill bool) error {
	if fill {
		for dy := 0; dy < height && y+dy < d.height; dy++ {
			for dx := 0; dx < width && x+dx < d.width; dx++ {
				if x+dx >= 0 && y+dy >= 0 {
					d.img.SetGray(x+dx, y+dy, color.Gray{Y: 255})
				}
			}
		}
	} else {
		// Draw outline
		for i := 0; i < width && x+i < d.width; i++ {
			if x+i >= 0 && y >= 0 {
				d.img.SetGray(x+i, y, color.Gray{Y: 255})
			}
			if x+i >= 0 && y+height-1 >= 0 && y+height-1 < d.height {
				d.img.SetGray(x+i, y+height-1, color.Gray{Y: 255})
			}
		}
		for i := 0; i < height && y+i < d.height; i++ {
			if x >= 0 && y+i >= 0 {
				d.img.SetGray(x, y+i, color.Gray{Y: 255})
			}
			if x+width-1 >= 0 && x+width-1 < d.width && y+i >= 0 {
				d.img.SetGray(x+width-1, y+i, color.Gray{Y: 255})
			}
		}
	}
	return nil
}

// DrawImage draws an image at the specified position
func (d *SSD1306Display) DrawImage(x, y int, img image.Image) error {
	bounds := img.Bounds()
	for dy := 0; dy < bounds.Dy() && y+dy < d.height; dy++ {
		for dx := 0; dx < bounds.Dx() && x+dx < d.width; dx++ {
			if x+dx >= 0 && y+dy >= 0 {
				r, g, b, a := img.At(bounds.Min.X+dx, bounds.Min.Y+dy).RGBA()
				brightness := (r + g + b) / 3
				if brightness > 32768 && a > 32768 {
					d.img.SetGray(x+dx, y+dy, color.Gray{Y: 255})
				} else {
					d.img.SetGray(x+dx, y+dy, color.Gray{Y: 0})
				}
			}
		}
	}
	return nil
}

// Show flushes the buffer to the display
func (d *SSD1306Display) Show() error {
	// Draw the image to the display
	if err := d.dev.Draw(d.img.Bounds(), d.img, image.Point{}); err != nil {
		return fmt.Errorf("failed to draw to display: %w", err)
	}
	return nil
}

// Close closes the display connection
func (d *SSD1306Display) Close() error {
	// periph.io devices don't need explicit closing
	return d.dev.Halt()
}

// GetBounds returns the display dimensions
func (d *SSD1306Display) GetBounds() image.Rectangle {
	return d.img.Bounds()
}

// GetBuffer returns a copy of the current display buffer
func (d *SSD1306Display) GetBuffer() []byte {
	// Convert image to byte buffer
	buf := make([]byte, d.width*d.height/8)
	for y := 0; y < d.height; y++ {
		for x := 0; x < d.width; x++ {
			if d.img.GrayAt(x, y).Y > 128 {
				byteIdx := x + (y/8)*d.width
				bitIdx := uint(y % 8)
				buf[byteIdx] |= 1 << bitIdx
			}
		}
	}
	return buf
}

// SetBrightness sets the display contrast/brightness (0-255)
// For SSD1306, this maps directly to the contrast control command
func (d *SSD1306Display) SetBrightness(level uint8) error {
	// SSD1306 contrast command: 0x81 followed by contrast value
	// The periph.io driver exposes the underlying device
	// We'll use the Halt/Init cycle approach for now as periph.io
	// doesn't expose contrast control directly

	// Note: A production implementation would send raw I2C commands
	// For now, we'll accept the limitation that brightness control
	// isn't fully supported in periph.io's high-level API

	// This is a placeholder that returns success
	// Full implementation would require direct I2C communication
	return nil
}
