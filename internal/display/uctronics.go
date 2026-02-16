package display

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
	"time"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

// UCTRONICS MCU I2C bridge protocol constants.
// The UCTRONICS Pi Rack Pro (SKU_RM0004) has an onboard MCU that bridges
// I2C to the ST7735 display via SPI internally.  The host communicates
// with the MCU using a simple register protocol over I2C.
const (
	uctronicsDefaultAddr uint16 = 0x18

	// MCU register addresses
	uctronicsWriteDataReg  byte = 0x00 // single-pixel write: [reg, high, low]
	uctronicsBurstWriteReg byte = 0x01 // burst mode control: [reg, 0x00, 0x01=on/0x00=off]
	uctronicsSyncReg       byte = 0x03 // sync/flush: [reg, 0x00, 0x01]

	// ST7735 register addresses forwarded by the MCU
	uctronicsXCoordReg   byte = 0x2A // CASET: [reg, x0+xstart, x1+xstart]
	uctronicsYCoordReg   byte = 0x2B // RASET: [reg, y0+ystart, y1+ystart]
	uctronicsCharDataReg byte = 0x2C // RAMWR: [reg, 0x00, 0x00]

	// Panel position in ST7735 controller RAM
	uctronicsXStart byte = 0
	uctronicsYStart byte = 24

	// Maximum bytes per I2C burst chunk (from UCTRONICS reference driver)
	uctronicsBurstMaxLen = 160
)

// UCTRONICSDisplay implements Display for UCTRONICS I2C-bridged ST7735 displays.
type UCTRONICSDisplay struct {
	bus    i2c.BusCloser
	addr   uint16
	img    *image.NRGBA
	width  int
	height int
}

// NewUCTRONICSDisplay creates a new UCTRONICS display driver.
func NewUCTRONICSDisplay(i2cBus, i2cAddr string, width, height int) (*UCTRONICSDisplay, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize periph: %w", err)
	}

	bus, err := i2creg.Open(i2cBus)
	if err != nil {
		return nil, fmt.Errorf("failed to open I2C bus %s: %w", i2cBus, err)
	}

	addr, err := parseI2CAddr(i2cAddr)
	if err != nil {
		bus.Close() //nolint:errcheck
		return nil, err
	}
	if addr == 0 {
		addr = uctronicsDefaultAddr
	}

	return &UCTRONICSDisplay{
		bus:    bus,
		addr:   addr,
		img:    image.NewNRGBA(image.Rect(0, 0, width, height)),
		width:  width,
		height: height,
	}, nil
}

// parseI2CAddr converts a hex string like "0x18" to a uint16 address.
func parseI2CAddr(s string) (uint16, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if strings.HasPrefix(s, "0x") {
		s = s[2:]
	}
	v, err := strconv.ParseUint(s, 16, 16)
	if err != nil {
		return 0, fmt.Errorf("invalid I2C address: %w", err)
	}
	return uint16(v), nil
}

// writeCommand sends a 3-byte command to the MCU: [register, high, low].
func (d *UCTRONICSDisplay) writeCommand(reg, high, low byte) error {
	if err := d.bus.Tx(d.addr, []byte{reg, high, low}, nil); err != nil {
		return err
	}
	time.Sleep(10 * time.Microsecond)
	return nil
}

// setAddressWindow sets the pixel coordinate window on the display.
func (d *UCTRONICSDisplay) setAddressWindow(x0, y0, x1, y1 byte) error {
	if err := d.writeCommand(uctronicsXCoordReg, x0+uctronicsXStart, x1+uctronicsXStart); err != nil {
		return err
	}
	if err := d.writeCommand(uctronicsYCoordReg, y0+uctronicsYStart, y1+uctronicsYStart); err != nil {
		return err
	}
	if err := d.writeCommand(uctronicsCharDataReg, 0x00, 0x00); err != nil {
		return err
	}
	return d.writeCommand(uctronicsSyncReg, 0x00, 0x01)
}

// burstTransfer sends pixel data in chunks using the MCU's burst protocol.
func (d *UCTRONICSDisplay) burstTransfer(data []byte) error {
	// Enable burst mode
	if err := d.writeCommand(uctronicsBurstWriteReg, 0x00, 0x01); err != nil {
		return fmt.Errorf("burst enable failed: %w", err)
	}

	for offset := 0; offset < len(data); {
		end := offset + uctronicsBurstMaxLen
		if end > len(data) {
			end = len(data)
		}
		if err := d.bus.Tx(d.addr, data[offset:end], nil); err != nil {
			return fmt.Errorf("burst write failed at offset %d: %w", offset, err)
		}
		offset = end
		time.Sleep(700 * time.Microsecond)
	}

	// Disable burst mode
	if err := d.writeCommand(uctronicsBurstWriteReg, 0x00, 0x00); err != nil {
		return fmt.Errorf("burst disable failed: %w", err)
	}
	// Sync
	return d.writeCommand(uctronicsSyncReg, 0x00, 0x01)
}

// Init initializes the display (MCU handles ST7735 init; we just clear).
func (d *UCTRONICSDisplay) Init() error {
	return d.Clear()
}

// Clear fills the display with black.
func (d *UCTRONICSDisplay) Clear() error {
	bounds := d.img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			d.img.SetNRGBA(x, y, color.NRGBA{A: 255})
		}
	}
	return d.Show()
}

// DrawPixel sets a single pixel (white if on, black if off).
func (d *UCTRONICSDisplay) DrawPixel(x, y int, on bool) error {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return nil
	}
	if on {
		d.img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	} else {
		d.img.SetNRGBA(x, y, color.NRGBA{A: 255})
	}
	return nil
}

// DrawLine draws a horizontal line.
func (d *UCTRONICSDisplay) DrawLine(x, y, width int) error {
	for i := 0; i < width && x+i < d.width; i++ {
		if x+i >= 0 && y >= 0 && y < d.height {
			d.img.SetNRGBA(x+i, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	return nil
}

// DrawText draws text as simple character outlines.
func (d *UCTRONICSDisplay) DrawText(x, y int, text string, size int) error {
	charWidth := size / 2
	for i := range text {
		startX := x + i*charWidth
		if startX >= d.width {
			break
		}
		if err := d.DrawRect(startX, y, charWidth-1, size, false); err != nil {
			return err
		}
	}
	return nil
}

// DrawRect draws a rectangle outline or filled rectangle.
//
//nolint:gocyclo // drawing logic naturally has many conditional branches
func (d *UCTRONICSDisplay) DrawRect(x, y, width, height int, fill bool) error {
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	if fill {
		for dy := 0; dy < height && y+dy < d.height; dy++ {
			for dx := 0; dx < width && x+dx < d.width; dx++ {
				if x+dx >= 0 && y+dy >= 0 {
					d.img.SetNRGBA(x+dx, y+dy, white)
				}
			}
		}
	} else {
		for i := 0; i < width && x+i < d.width; i++ {
			if x+i >= 0 && y >= 0 {
				d.img.SetNRGBA(x+i, y, white)
			}
			if x+i >= 0 && y+height-1 >= 0 && y+height-1 < d.height {
				d.img.SetNRGBA(x+i, y+height-1, white)
			}
		}
		for i := 0; i < height && y+i < d.height; i++ {
			if x >= 0 && y+i >= 0 {
				d.img.SetNRGBA(x, y+i, white)
			}
			if x+width-1 >= 0 && x+width-1 < d.width && y+i >= 0 {
				d.img.SetNRGBA(x+width-1, y+i, white)
			}
		}
	}
	return nil
}

// DrawImage draws an image at the specified position, preserving source colours.
func (d *UCTRONICSDisplay) DrawImage(x, y int, img image.Image) error {
	bounds := img.Bounds()
	for dy := 0; dy < bounds.Dy() && y+dy < d.height; dy++ {
		for dx := 0; dx < bounds.Dx() && x+dx < d.width; dx++ {
			if x+dx >= 0 && y+dy >= 0 {
				r, g, b, a := img.At(bounds.Min.X+dx, bounds.Min.Y+dy).RGBA()
				if a > 32768 {
					d.img.SetNRGBA(x+dx, y+dy, color.NRGBA{
						R: uint8(r >> 8),
						G: uint8(g >> 8),
						B: uint8(b >> 8),
						A: 255,
					})
				} else {
					d.img.SetNRGBA(x+dx, y+dy, color.NRGBA{A: 255})
				}
			}
		}
	}
	return nil
}

// Show flushes the NRGBA buffer to the display as RGB565 via I2C burst transfer.
func (d *UCTRONICSDisplay) Show() error {
	if err := d.setAddressWindow(0, 0, byte(d.width-1), byte(d.height-1)); err != nil {
		return err
	}

	buf := make([]byte, d.width*d.height*2)
	idx := 0
	for y := 0; y < d.height; y++ {
		for x := 0; x < d.width; x++ {
			c := d.img.NRGBAAt(x, y)
			rgb565 := nrgbaToRGB565(c)
			buf[idx] = byte(rgb565 >> 8)
			buf[idx+1] = byte(rgb565)
			idx += 2
		}
	}

	return d.burstTransfer(buf)
}

// Close closes the I2C bus.
func (d *UCTRONICSDisplay) Close() error {
	return d.bus.Close()
}

// GetBounds returns the display dimensions.
func (d *UCTRONICSDisplay) GetBounds() image.Rectangle {
	return d.img.Bounds()
}

// GetBuffer returns the current frame as RGB565-encoded bytes.
func (d *UCTRONICSDisplay) GetBuffer() []byte {
	buf := make([]byte, d.width*d.height*2)
	idx := 0
	for y := 0; y < d.height; y++ {
		for x := 0; x < d.width; x++ {
			c := d.img.NRGBAAt(x, y)
			rgb565 := nrgbaToRGB565(c)
			buf[idx] = byte(rgb565 >> 8)
			buf[idx+1] = byte(rgb565)
			idx += 2
		}
	}
	return buf
}

// SetBrightness is a no-op (UCTRONICS MCU does not expose brightness control).
func (d *UCTRONICSDisplay) SetBrightness(_ uint8) error {
	return nil
}
