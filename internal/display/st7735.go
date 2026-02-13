package display

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"
)

// ST7735 command bytes
const (
	st7735SWRESET = 0x01
	st7735SLPOUT  = 0x11
	st7735NORON   = 0x13
	st7735DISPON  = 0x29
	st7735CASET   = 0x2A
	st7735RASET   = 0x2B
	st7735RAMWR   = 0x2C
	st7735COLMOD  = 0x3A
	st7735MADCTL  = 0x36
	st7735FRMCTR1 = 0xB1
	st7735FRMCTR2 = 0xB2
	st7735FRMCTR3 = 0xB3
	st7735INVCTR  = 0xB4
	st7735PWCTR1  = 0xC0
	st7735PWCTR2  = 0xC1
	st7735PWCTR3  = 0xC2
	st7735PWCTR4  = 0xC3
	st7735PWCTR5  = 0xC4
	st7735VMCTR1  = 0xC5
	st7735GMCTRP1 = 0xE0
	st7735GMCTRN1 = 0xE1
)

// MADCTL flags
const (
	madctlMY  = 0x80
	madctlMX  = 0x40
	madctlMV  = 0x20
	madctlBGR = 0x08
)

// ST7735Display implements Display interface for ST7735 TFT displays via SPI
type ST7735Display struct {
	port      spi.PortCloser
	conn      spi.Conn
	dc        gpio.PinOut
	rst       gpio.PinOut // nil if not configured
	img       *image.NRGBA
	width     int
	height    int
	colOffset uint8
	rowOffset uint8
}

// NewST7735Display creates a new ST7735 display driver
func NewST7735Display(spiBus, dcPin, rstPin string, width, height, rotation int) (*ST7735Display, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize periph: %w", err)
	}

	port, err := spireg.Open(spiBus)
	if err != nil {
		return nil, fmt.Errorf("failed to open SPI bus %s: %w", spiBus, err)
	}

	conn, err := port.Connect(15*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		port.Close() //nolint:errcheck
		return nil, fmt.Errorf("failed to connect on SPI bus %s: %w", spiBus, err)
	}

	dc := gpioreg.ByName(dcPin)
	if dc == nil {
		port.Close() //nolint:errcheck
		return nil, fmt.Errorf("DC pin %q not found", dcPin)
	}

	var rst gpio.PinOut
	if rstPin != "" {
		rst = gpioreg.ByName(rstPin)
		if rst == nil {
			port.Close() //nolint:errcheck
			return nil, fmt.Errorf("RST pin %q not found", rstPin)
		}
	}

	colOffset, rowOffset := st7735Offsets(width, height)

	d := &ST7735Display{
		port:      port,
		conn:      conn,
		dc:        dc,
		rst:       rst,
		img:       image.NewNRGBA(image.Rect(0, 0, width, height)),
		width:     width,
		height:    height,
		colOffset: colOffset,
		rowOffset: rowOffset,
	}

	if err := d.hardwareReset(); err != nil {
		port.Close() //nolint:errcheck
		return nil, err
	}

	if err := d.initSequence(); err != nil {
		port.Close() //nolint:errcheck
		return nil, err
	}

	if err := d.applyRotation(rotation); err != nil {
		port.Close() //nolint:errcheck
		return nil, err
	}

	return d, nil
}

// st7735Offsets returns the column and row offsets for the given display dimensions.
// Different ST7735 panel variants require pixel offsets to address the correct
// area of the controller's internal RAM.
func st7735Offsets(width, height int) (colOffset, rowOffset uint8) {
	switch {
	case width == 160 && height == 80:
		return 0, 24
	case width == 128 && height == 128:
		return 2, 3
	default: // 128x160
		return 0, 0
	}
}

func (d *ST7735Display) hardwareReset() error {
	if d.rst == nil {
		return nil
	}
	if err := d.rst.Out(gpio.High); err != nil {
		return fmt.Errorf("RST high failed: %w", err)
	}
	time.Sleep(5 * time.Millisecond)
	if err := d.rst.Out(gpio.Low); err != nil {
		return fmt.Errorf("RST low failed: %w", err)
	}
	time.Sleep(20 * time.Millisecond)
	if err := d.rst.Out(gpio.High); err != nil {
		return fmt.Errorf("RST high failed: %w", err)
	}
	time.Sleep(150 * time.Millisecond)
	return nil
}

//nolint:gocyclo // init sequence naturally has many steps
func (d *ST7735Display) initSequence() error {
	seq := []func() error{
		func() error { return d.sendCmd(st7735SWRESET) },
		func() error { time.Sleep(150 * time.Millisecond); return nil },
		func() error { return d.sendCmd(st7735SLPOUT) },
		func() error { time.Sleep(500 * time.Millisecond); return nil },
		func() error { return d.sendCmdData(st7735FRMCTR1, 0x01, 0x2C, 0x2D) },
		func() error { return d.sendCmdData(st7735FRMCTR2, 0x01, 0x2C, 0x2D) },
		func() error {
			return d.sendCmdData(st7735FRMCTR3, 0x01, 0x2C, 0x2D, 0x01, 0x2C, 0x2D)
		},
		func() error { return d.sendCmdData(st7735INVCTR, 0x07) },
		func() error { return d.sendCmdData(st7735PWCTR1, 0xA2, 0x02, 0x84) },
		func() error { return d.sendCmdData(st7735PWCTR2, 0xC5) },
		func() error { return d.sendCmdData(st7735PWCTR3, 0x0A, 0x00) },
		func() error { return d.sendCmdData(st7735PWCTR4, 0x8A, 0x2A) },
		func() error { return d.sendCmdData(st7735PWCTR5, 0x8A, 0xEE) },
		func() error { return d.sendCmdData(st7735VMCTR1, 0x0E) },
		func() error { return d.sendCmdData(st7735COLMOD, 0x05) }, // RGB565
		func() error {
			return d.sendCmdData(st7735GMCTRP1,
				0x02, 0x1C, 0x07, 0x12, 0x37, 0x32, 0x29, 0x2D,
				0x29, 0x25, 0x2B, 0x39, 0x00, 0x01, 0x03, 0x10)
		},
		func() error {
			return d.sendCmdData(st7735GMCTRN1,
				0x03, 0x1D, 0x07, 0x06, 0x2E, 0x2C, 0x29, 0x2D,
				0x2E, 0x2E, 0x37, 0x3F, 0x00, 0x00, 0x02, 0x10)
		},
		func() error { return d.sendCmd(st7735NORON) },
		func() error { return d.sendCmd(st7735DISPON) },
		func() error { time.Sleep(100 * time.Millisecond); return nil },
	}

	for _, step := range seq {
		if err := step(); err != nil {
			return fmt.Errorf("ST7735 init sequence failed: %w", err)
		}
	}
	return nil
}

func (d *ST7735Display) applyRotation(rotation int) error {
	var madctl byte
	switch rotation {
	case 0:
		madctl = madctlMX | madctlMY
	case 1:
		madctl = madctlMY | madctlMV
	case 2:
		madctl = 0x00
	case 3:
		madctl = madctlMX | madctlMV
	default:
		return fmt.Errorf("ST7735 rotation must be 0-3, got %d", rotation)
	}
	return d.sendCmdData(st7735MADCTL, madctl)
}

// sendCmd asserts DC low and transmits a single command byte.
func (d *ST7735Display) sendCmd(cmd byte) error {
	if err := d.dc.Out(gpio.Low); err != nil {
		return err
	}
	return d.conn.Tx([]byte{cmd}, nil)
}

// spiMaxTx is the maximum number of bytes per SPI transaction on sysfs.
const spiMaxTx = 4096

// sendData asserts DC high and transmits data bytes, chunking as needed
// to respect the sysfs SPI driver's 4096-byte per-transaction limit.
func (d *ST7735Display) sendData(data ...byte) error {
	if err := d.dc.Out(gpio.High); err != nil {
		return err
	}
	for len(data) > 0 {
		chunk := data
		if len(chunk) > spiMaxTx {
			chunk = data[:spiMaxTx]
		}
		if err := d.conn.Tx(chunk, nil); err != nil {
			return err
		}
		data = data[len(chunk):]
	}
	return nil
}

// sendCmdData sends a command followed by data bytes.
func (d *ST7735Display) sendCmdData(cmd byte, data ...byte) error {
	if err := d.sendCmd(cmd); err != nil {
		return err
	}
	if len(data) > 0 {
		return d.sendData(data...)
	}
	return nil
}

// setWindow sets the address window for subsequent RAMWR pixel data.
func (d *ST7735Display) setWindow(x0, y0, x1, y1 int) error {
	cx0 := uint8(x0) + d.colOffset
	cx1 := uint8(x1) + d.colOffset
	ry0 := uint8(y0) + d.rowOffset
	ry1 := uint8(y1) + d.rowOffset

	if err := d.sendCmdData(st7735CASET, 0x00, cx0, 0x00, cx1); err != nil {
		return err
	}
	if err := d.sendCmdData(st7735RASET, 0x00, ry0, 0x00, ry1); err != nil {
		return err
	}
	return d.sendCmd(st7735RAMWR)
}

// Init initializes the display (already done in constructor; clears screen).
func (d *ST7735Display) Init() error {
	return d.Clear()
}

// Clear fills the display with black.
func (d *ST7735Display) Clear() error {
	bounds := d.img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			d.img.SetNRGBA(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}
	return d.Show()
}

// DrawPixel sets a single pixel (white if on, black if off).
func (d *ST7735Display) DrawPixel(x, y int, on bool) error {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return nil
	}
	if on {
		d.img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	} else {
		d.img.SetNRGBA(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
	}
	return nil
}

// DrawLine draws a horizontal line.
func (d *ST7735Display) DrawLine(x, y, width int) error {
	for i := 0; i < width && x+i < d.width; i++ {
		if x+i >= 0 && y >= 0 && y < d.height {
			d.img.SetNRGBA(x+i, y, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	return nil
}

// DrawText draws text as simple character outlines.
func (d *ST7735Display) DrawText(x, y int, text string, size int) error {
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
func (d *ST7735Display) DrawRect(x, y, width, height int, fill bool) error {
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

// DrawImage draws an image at the specified position (thresholded to white/black).
func (d *ST7735Display) DrawImage(x, y int, img image.Image) error {
	bounds := img.Bounds()
	for dy := 0; dy < bounds.Dy() && y+dy < d.height; dy++ {
		for dx := 0; dx < bounds.Dx() && x+dx < d.width; dx++ {
			if x+dx >= 0 && y+dy >= 0 {
				r, g, b, a := img.At(bounds.Min.X+dx, bounds.Min.Y+dy).RGBA()
				brightness := (r + g + b) / 3
				if brightness > 32768 && a > 32768 {
					d.img.SetNRGBA(x+dx, y+dy, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
				} else {
					d.img.SetNRGBA(x+dx, y+dy, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
				}
			}
		}
	}
	return nil
}

// Show flushes the NRGBA buffer to the display as RGB565.
func (d *ST7735Display) Show() error {
	if err := d.setWindow(0, 0, d.width-1, d.height-1); err != nil {
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

	return d.sendData(buf...)
}

// nrgbaToRGB565 converts an NRGBA colour to a 16-bit RGB565 value.
func nrgbaToRGB565(c color.NRGBA) uint16 {
	r := uint16(c.R) >> 3
	g := uint16(c.G) >> 2
	b := uint16(c.B) >> 3
	return (r << 11) | (g << 5) | b
}

// Close closes the SPI port.
func (d *ST7735Display) Close() error {
	return d.port.Close()
}

// GetBounds returns the display dimensions.
func (d *ST7735Display) GetBounds() image.Rectangle {
	return d.img.Bounds()
}

// GetBuffer returns the current frame as RGB565-encoded bytes.
func (d *ST7735Display) GetBuffer() []byte {
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

// SetBrightness is a no-op placeholder (backlight control not in scope).
func (d *ST7735Display) SetBrightness(_ uint8) error {
	return nil
}
