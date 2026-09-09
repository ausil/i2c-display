package display

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/host/v3"
)

// ssd1306DefaultAddr is the I2C address hardcoded by periph's ssd1306 driver.
const ssd1306DefaultAddr uint16 = 0x3C

// SSD1306 command bytes used for brightness control.
const (
	ssd1306CmdPrefix   byte = 0x00 // control byte: the following bytes are commands
	ssd1306SetContrast byte = 0x81
	ssd1306DisplayOff  byte = 0xAE
	ssd1306DisplayOn   byte = 0xAF
)

// remappedI2CBus rewrites transactions addressed to `from` so they go to `to`.
// It works around drivers that hardcode the device address.
type remappedI2CBus struct {
	i2c.Bus
	from, to uint16
}

func (b *remappedI2CBus) Tx(addr uint16, w, r []byte) error {
	if addr == b.from {
		addr = b.to
	}
	return b.Bus.Tx(addr, w, r)
}

// SSD1306Display implements Display interface for real SSD1306 hardware
type SSD1306Display struct {
	dev    *ssd1306.Dev
	bus    i2c.Bus // raw bus, for commands periph's driver does not expose
	addr   uint16  // effective device address
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
		bus.Close() // #nosec G104 -- best-effort cleanup on error path
		return nil, fmt.Errorf("SSD1306 only supports rotation 0 (0°) and 2 (180°), got %d", rotation)
	}

	addr, err := parseI2CAddr(i2cAddr)
	if err != nil {
		bus.Close() // #nosec G104 -- best-effort cleanup on error path
		return nil, err
	}

	// periph's ssd1306.NewI2C hardcodes address 0x3C; remap transactions
	// when the display is strapped to a different address (commonly 0x3D).
	var devBus i2c.Bus = bus
	if addr != 0 && addr != ssd1306DefaultAddr {
		devBus = &remappedI2CBus{Bus: bus, from: ssd1306DefaultAddr, to: addr}
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
	dev, err := ssd1306.NewI2C(devBus, &opts)
	if err != nil {
		bus.Close() // #nosec G104 -- best-effort cleanup on error path
		return nil, fmt.Errorf("failed to create SSD1306 device: %w", err)
	}

	effectiveAddr := addr
	if effectiveAddr == 0 {
		effectiveAddr = ssd1306DefaultAddr
	}

	return &SSD1306Display{
		dev:    dev,
		bus:    bus,
		addr:   effectiveAddr,
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
//
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
			if x+dx < 0 || y+dy < 0 {
				continue
			}
			r, g, b, a := img.At(bounds.Min.X+dx, bounds.Min.Y+dy).RGBA()
			// Use max channel so saturated colours (e.g. pure green)
			// render as white on the monochrome display.
			brightness := r
			if g > brightness {
				brightness = g
			}
			if b > brightness {
				brightness = b
			}
			if brightness > 32768 && a > 32768 {
				d.img.SetGray(x+dx, y+dy, color.Gray{Y: 255})
			} else {
				d.img.SetGray(x+dx, y+dy, color.Gray{Y: 0})
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
				bitIdx := uint(y % 8) /* #nosec G115 -- modulo 8 is always 0–7 */
				buf[byteIdx] |= 1 << bitIdx
			}
		}
	}
	return buf
}

// SetBrightness sets the display brightness (0-255). Level 0 turns the panel
// off entirely (the screensaver's blank mode); any other level turns it back
// on and sets the contrast. periph's driver does not expose these commands,
// so they are sent as raw I2C writes. Each call is a single bus transaction
// and neither command touches the driver's addressing state, so this is safe
// alongside concurrent frame draws.
func (d *SSD1306Display) SetBrightness(level uint8) error {
	var cmds []byte
	if level == 0 {
		cmds = []byte{ssd1306CmdPrefix, ssd1306DisplayOff}
	} else {
		cmds = []byte{ssd1306CmdPrefix, ssd1306DisplayOn, ssd1306SetContrast, level}
	}
	if err := d.bus.Tx(d.addr, cmds, nil); err != nil {
		return fmt.Errorf("failed to set brightness: %w", err)
	}
	return nil
}
