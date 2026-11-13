package display

import (
	"fmt"

	"periph.io/x/conn/v3/i2c"
)

// SH1106 controller constants.
//
// The SH1106 is close to the SSD1306 but not drop-in compatible, which is why
// periph's SSD1306 driver cannot drive it:
//
//   - It only supports page addressing mode. The SSD1306's horizontal/vertical
//     addressing command (0x20) and the column/page range windows (0x21, 0x22)
//     are absent, so a frame has to be written page by page with an explicit
//     column pointer.
//   - Its GDDRAM is 132 columns wide while the panel drives 128 of them, so
//     column data has to be written with an offset of 2. Writing at column 0
//     (what an SSD1306 driver does) shifts the image two pixels left and wraps
//     the stale contents of columns 0-1 onto the right edge.
//   - The charge-pump command the SSD1306 uses (0x8D) does not exist here; the
//     SH1106 needs 0xAD 0x8B to enable its internal DC-DC converter. Without it
//     the panel stays completely dark.
const (
	sh1106DefaultAddr uint16 = 0x3C

	// sh1106ColumnOffset is the shift between panel pixels and GDDRAM columns.
	sh1106ColumnOffset = 2
	// sh1106Columns and sh1106Rows bound the controller's addressable RAM.
	sh1106Columns = 132
	sh1106Rows    = 64

	sh1106PageStart byte = 0xB0 // | page index
	sh1106ColLow    byte = 0x00 // | low nibble of the column address
	sh1106ColHigh   byte = 0x10 // | high nibble of the column address
)

// SH1106Display implements Display for SH1106 128x64 monochrome OLED panels,
// the controller found on most 1.3" (and 2.42") I2C OLED modules.
type SH1106Display struct {
	monoFrameBuffer
	bus    i2c.Bus
	closer i2c.BusCloser // non-nil only when this driver owns the bus
	addr   uint16
	rotate int // 0 or 2 (180°), applied via the mirror commands
}

// NewSH1106Display creates a new SH1106 display driver.
func NewSH1106Display(i2cBus, i2cAddr string, width, height, rotation int) (*SH1106Display, error) {
	bus, addr, err := openI2CBus(i2cBus, i2cAddr, sh1106DefaultAddr)
	if err != nil {
		return nil, err
	}

	d, err := newSH1106OnBus(bus, addr, width, height, rotation)
	if err != nil {
		bus.Close() // #nosec G104 -- best-effort cleanup on error path
		return nil, err
	}
	d.closer = bus
	return d, nil
}

// newSH1106OnBus validates the geometry, brings the panel up on an already
// opened bus and returns the driver. Split out from NewSH1106Display so the bus
// traffic can be exercised against a fake bus in tests.
func newSH1106OnBus(bus i2c.Bus, addr uint16, width, height, rotation int) (*SH1106Display, error) {
	if rotation != 0 && rotation != 2 {
		return nil, fmt.Errorf("SH1106 only supports rotation 0 (0°) and 2 (180°), got %d", rotation)
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("SH1106 needs positive dimensions, got %dx%d", width, height)
	}
	if height%8 != 0 {
		return nil, fmt.Errorf("SH1106 height must be a multiple of 8 (page size), got %d", height)
	}
	if height > sh1106Rows {
		return nil, fmt.Errorf("SH1106 drives at most %d rows, got %d", sh1106Rows, height)
	}
	if width+sh1106ColumnOffset > sh1106Columns {
		return nil, fmt.Errorf("SH1106 GDDRAM is %d columns wide and reserves %d as an offset, so width must be at most %d, got %d",
			sh1106Columns, sh1106ColumnOffset, sh1106Columns-sh1106ColumnOffset, width)
	}

	d := &SH1106Display{
		monoFrameBuffer: newMonoFrameBuffer(width, height),
		bus:             bus,
		addr:            addr,
		rotate:          rotation,
	}

	if err := d.sendCommands(sh1106InitSeq(rotation)); err != nil {
		return nil, fmt.Errorf("failed to initialize SH1106: %w", err)
	}
	return d, nil
}

// sh1106InitSeq returns the power-on sequence for an I2C SH1106 panel.
//
// Rotation is done with the mirror commands rather than in software: segment
// remap flips the image horizontally, COM scan direction flips it vertically,
// and setting both is a 180° rotation.
func sh1106InitSeq(rotation int) []byte {
	segmentRemap := byte(0xA1) // column 127 -> SEG127
	comScan := byte(0xC8)      // COM127 -> COM0
	if rotation == 2 {
		segmentRemap = 0xA0
		comScan = 0xC0
	}

	return []byte{
		0xAE,       // display off
		0xD5, 0x80, // clock divide ratio and oscillator frequency
		0xA8, 0x3F, // multiplex ratio: 64 rows
		0xD3, 0x00, // display offset: 0
		0x40,       // display start line: 0
		0xAD, 0x8B, // enable the internal DC-DC power circuit
		0x81, 0x80, // contrast
		0xA4, // display follows GDDRAM content
		0xA6, // normal display (not inverted)
		segmentRemap,
		comScan,
		0xDA, 0x12, // COM pins: alternative configuration (128x64 panel)
		0xD9, 0x1F, // pre-charge period
		0xDB, 0x30, // VCOMH deselect level
		0xAF, // display on
	}
}

// Init re-applies the power-on configuration and clears the panel so it does
// not show garbage left in GDDRAM by the bootloader or a previous run.
func (d *SH1106Display) Init() error {
	if err := d.sendCommands(sh1106InitSeq(d.rotate)); err != nil {
		return fmt.Errorf("failed to initialize SH1106: %w", err)
	}
	if err := d.Clear(); err != nil {
		return err
	}
	return d.Show()
}

// Show flushes the buffer to the display, one transaction pair per 8-row page:
// set the page and column pointer, then stream that page's GDDRAM bytes.
func (d *SH1106Display) Show() error {
	for page := range d.pages() {
		addrCmds := []byte{
			sh1106PageStart | byte(page), // #nosec G115 -- page index is bounded by the height check
			sh1106ColLow | byte(sh1106ColumnOffset&0x0F),
			sh1106ColHigh | byte((sh1106ColumnOffset>>4)&0x07),
		}
		if err := d.sendCommands(addrCmds); err != nil {
			return fmt.Errorf("failed to set page %d address: %w", page, err)
		}

		data := make([]byte, 0, 1+d.width)
		data = append(data, monoDataPrefix)
		data = append(data, d.pageBytes(page)...)
		if err := d.bus.Tx(d.addr, data, nil); err != nil {
			return fmt.Errorf("failed to write page %d: %w", page, err)
		}
	}
	return nil
}

// Close turns the panel off and releases the bus.
func (d *SH1106Display) Close() error {
	err := d.sendCommands([]byte{monoDisplayOff})
	if d.closer == nil {
		return err
	}
	if closeErr := d.closer.Close(); err == nil {
		err = closeErr
	}
	return err
}

// SetBrightness sets the display brightness (0-255). Level 0 turns the panel
// off entirely (the screensaver's blank mode); any other level turns it back on
// and sets the contrast.
func (d *SH1106Display) SetBrightness(level uint8) error {
	return setMonoBrightness(d.bus, d.addr, level)
}

// sendCommands writes command bytes behind the command control byte.
func (d *SH1106Display) sendCommands(cmds []byte) error {
	w := make([]byte, 0, len(cmds)+1)
	w = append(w, monoCmdPrefix)
	w = append(w, cmds...)
	if err := d.bus.Tx(d.addr, w, nil); err != nil {
		return fmt.Errorf("sh1106 command failed: %w", err)
	}
	return nil
}
