package display

import (
	"fmt"
	"image"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/host/v3"
)

// ssd1306DefaultAddr is the I2C address hardcoded by periph's ssd1306 driver.
const ssd1306DefaultAddr uint16 = 0x3C

// SSD1306 command bytes used for brightness control. The controller shares the
// SSD1306/SH1106 command encoding, so these alias the mono constants.
const (
	ssd1306CmdPrefix   = monoCmdPrefix
	ssd1306SetContrast = monoSetContrast
	ssd1306DisplayOff  = monoDisplayOff
	ssd1306DisplayOn   = monoDisplayOn
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
	monoFrameBuffer
	dev  *ssd1306.Dev
	bus  i2c.Bus // raw bus, for commands periph's driver does not expose
	addr uint16  // effective device address
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
		monoFrameBuffer: newMonoFrameBuffer(width, height),
		dev:             dev,
		bus:             bus,
		addr:            effectiveAddr,
	}, nil
}

// Init initializes the display
func (d *SSD1306Display) Init() error {
	// The device is initialized in NewSSD1306Display
	// Clear the display to start fresh
	return d.Clear()
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

// SetBrightness sets the display brightness (0-255). Level 0 turns the panel
// off entirely (the screensaver's blank mode); any other level turns it back
// on and sets the contrast. periph's driver does not expose these commands,
// so they are sent as raw I2C writes.
func (d *SSD1306Display) SetBrightness(level uint8) error {
	return setMonoBrightness(d.bus, d.addr, level)
}
