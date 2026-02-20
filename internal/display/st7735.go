package display

import (
	"fmt"
	"image"
	"image/color"
	"log"
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
	madctlML  = 0x10
	madctlBGR = 0x08
)

// ST7735Display implements Display interface for ST7735 TFT displays via SPI
type ST7735Display struct {
	port        spi.PortCloser
	conn        spi.Conn
	dc          gpio.PinOut
	rst         gpio.PinOut // nil if not configured
	img         *image.NRGBA
	width       int
	height      int
	panelWidth  int    // physical panel width (before rotation)
	panelHeight int    // physical panel height (before rotation)
	displayType string // full display type name for variant-specific behaviour
	colOffset   uint8
	rowOffset   uint8
}

// NewST7735Display creates a new ST7735 display driver
//
//nolint:gocyclo // initialization naturally has many sequential error-checked steps
func NewST7735Display(spiBus, dcPin, rstPin string, width, height, rotation int, displayType string) (*ST7735Display, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize periph: %w", err)
	}

	port, err := spireg.Open(spiBus)
	if err != nil {
		return nil, fmt.Errorf("failed to open SPI bus %s: %w", spiBus, err)
	}

	conn, err := port.Connect(15*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		if cerr := port.Close(); cerr != nil {
			log.Printf("st7735: failed to close SPI port during cleanup: %v", cerr)
		}
		return nil, fmt.Errorf("failed to connect on SPI bus %s: %w", spiBus, err)
	}

	dc := gpioreg.ByName(dcPin)
	if dc == nil {
		if cerr := port.Close(); cerr != nil {
			log.Printf("st7735: failed to close SPI port during cleanup: %v", cerr)
		}
		return nil, fmt.Errorf("DC pin %q not found", dcPin)
	}

	var rst gpio.PinOut
	if rstPin != "" {
		rst = gpioreg.ByName(rstPin)
		if rst == nil {
			if cerr := port.Close(); cerr != nil {
				log.Printf("st7735: failed to close SPI port during cleanup: %v", cerr)
			}
			return nil, fmt.Errorf("RST pin %q not found", rstPin)
		}
	}

	d := &ST7735Display{
		port:        port,
		conn:        conn,
		dc:          dc,
		rst:         rst,
		img:         image.NewNRGBA(image.Rect(0, 0, width, height)),
		width:       width,
		height:      height,
		panelWidth:  width,
		panelHeight: height,
		displayType: displayType,
	}

	if err := d.hardwareReset(); err != nil {
		if cerr := port.Close(); cerr != nil {
			log.Printf("st7735: failed to close SPI port during cleanup: %v", cerr)
		}
		return nil, err
	}

	if err := d.initSequence(); err != nil {
		if cerr := port.Close(); cerr != nil {
			log.Printf("st7735: failed to close SPI port during cleanup: %v", cerr)
		}
		return nil, err
	}

	if err := d.applyRotation(rotation); err != nil {
		if cerr := port.Close(); cerr != nil {
			log.Printf("st7735: failed to close SPI port during cleanup: %v", cerr)
		}
		return nil, err
	}

	return d, nil
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
	madctl, colOff, rowOff := d.st7735RotationParams(rotation)
	if rotation < 0 || rotation > 3 {
		return fmt.Errorf("ST7735 rotation must be 0-3, got %d", rotation)
	}
	d.colOffset = colOff
	d.rowOffset = rowOff
	return d.sendCmdData(st7735MADCTL, madctl)
}

// st7735RotationParams returns the MADCTL byte and RAM offsets for a given
// rotation.  The 160x80 panel is special: the ST7735 controller has 132
// columns × 162 rows of RAM, so the 160-pixel dimension MUST be mapped to
// the row axis via the MV (row/column exchange) bit.
func (d *ST7735Display) st7735RotationParams(rotation int) (madctl byte, colOffset, rowOffset uint8) {
	if d.panelWidth == 160 && d.panelHeight == 80 {
		// MADCTL values confirmed against Waveshare 0.96" 160x80 reference driver.
		// MV must be set for all landscape orientations so the 160-pixel dimension
		// maps to the controller's 162-row axis (max columns is only 132).
		switch rotation {
		case 0: // landscape normal — Waveshare reference: 0x70
			return madctlMX | madctlMV | madctlML, 1, 26
		case 1: // landscape 90° CW
			return madctlMY | madctlMV | madctlML, 26, 1
		case 2: // landscape 180°
			return madctlMY | madctlMV, 1, 26
		default: // landscape 270° CW
			return madctlMX | madctlMV, 26, 1
		}
	}

	if d.panelWidth == 128 && d.panelHeight == 128 {
		switch rotation {
		case 0:
			return madctlMX | madctlMY, 2, 3
		case 1:
			return madctlMY | madctlMV, 3, 2
		case 2:
			return 0x00, 2, 1
		default:
			return madctlMX | madctlMV, 1, 2
		}
	}

	// 128x160 — uses the full RAM, no offset needed
	switch rotation {
	case 0:
		return madctlMX | madctlMY, 0, 0
	case 1:
		return madctlMY | madctlMV, 0, 0
	case 2:
		return 0x00, 0, 0
	default:
		return madctlMX | madctlMV, 0, 0
	}
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
	cx0 := uint8(x0) + d.colOffset /* #nosec G115 -- display coordinates bounded by ≤255 dimensions */
	cx1 := uint8(x1) + d.colOffset /* #nosec G115 -- display coordinates bounded by ≤255 dimensions */
	ry0 := uint8(y0) + d.rowOffset /* #nosec G115 -- display coordinates bounded by ≤255 dimensions */
	ry1 := uint8(y1) + d.rowOffset /* #nosec G115 -- display coordinates bounded by ≤255 dimensions */

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
	if err := d.Clear(); err != nil {
		return err
	}
	return d.Show()
}

// Clear fills the image buffer with black without flushing to the display.
func (d *ST7735Display) Clear() error {
	bounds := d.img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			d.img.SetNRGBA(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}
	return nil
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
func (d *ST7735Display) DrawRect(x, y, width, height int, fill bool) error {
	drawRectNRGBA(d.img, x, y, width, height, d.width, d.height, fill)
	return nil
}

// DrawImage draws an image at the specified position, preserving source colours.
func (d *ST7735Display) DrawImage(x, y int, img image.Image) error {
	drawImageNRGBA(d.img, x, y, d.width, d.height, img)
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
			buf[idx] = byte(rgb565 >> 8) // #nosec G115 -- uint16 to byte truncation is intentional
			buf[idx+1] = byte(rgb565)    // #nosec G115 -- uint16 to byte truncation is intentional
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
			buf[idx] = byte(rgb565 >> 8) // #nosec G115 -- uint16 to byte truncation is intentional
			buf[idx+1] = byte(rgb565)    // #nosec G115 -- uint16 to byte truncation is intentional
			idx += 2
		}
	}
	return buf
}

// SetBrightness is a no-op placeholder (backlight control not in scope).
func (d *ST7735Display) SetBrightness(_ uint8) error {
	return nil
}
