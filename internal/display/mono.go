package display

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"periph.io/x/conn/v3/i2c"
)

// luminanceThreshold splits source pixels into lit/unlit when compositing
// colour artwork onto a monochrome panel.
const luminanceThreshold = 32768

// I2C control bytes and commands shared by the SSD1306 and SH1106 families.
// Both controllers take a control byte in front of every transfer and agree on
// these command encodings.
const (
	monoCmdPrefix   byte = 0x00 // control byte: the following bytes are commands
	monoDataPrefix  byte = 0x40 // control byte: the following bytes are GDDRAM data
	monoSetContrast byte = 0x81
	monoDisplayOff  byte = 0xAE
	monoDisplayOn   byte = 0xAF
)

// setMonoBrightness turns the panel on or off and sets its contrast.
//
// Level 0 turns the panel off entirely (the screensaver's blank mode); any
// other level turns it back on and applies the contrast. Each call is a single
// bus transaction and neither command touches the controller's addressing
// state, so this is safe alongside concurrent frame draws.
func setMonoBrightness(bus i2c.Bus, addr uint16, level uint8) error {
	var cmds []byte
	if level == 0 {
		cmds = []byte{monoCmdPrefix, monoDisplayOff}
	} else {
		cmds = []byte{monoCmdPrefix, monoDisplayOn, monoSetContrast, level}
	}
	if err := bus.Tx(addr, cmds, nil); err != nil {
		return fmt.Errorf("failed to set brightness: %w", err)
	}
	return nil
}

// monoFrameBuffer is a 1-bit-per-pixel framebuffer (held as 8-bit gray, where
// any non-zero value means "lit") plus the drawing primitives shared by the
// monochrome OLED controllers, SSD1306 and SH1106.
//
// Embedding types provide the transport half (Init, Show, Close,
// SetBrightness) that pushes this buffer to the hardware.
type monoFrameBuffer struct {
	img    *image.Gray
	width  int
	height int
}

// newMonoFrameBuffer returns a blank frame buffer for a width x height panel.
func newMonoFrameBuffer(width, height int) monoFrameBuffer {
	return monoFrameBuffer{
		img:    image.NewGray(image.Rect(0, 0, width, height)),
		width:  width,
		height: height,
	}
}

// Clear blanks the buffer without flushing it to the display.
func (m *monoFrameBuffer) Clear() error {
	draw.Draw(m.img, m.img.Bounds(), &image.Uniform{color.Gray{Y: 0}}, image.Point{}, draw.Src)
	return nil
}

// DrawText draws text at the specified position.
//
// The monochrome panels have no glyph renderer of their own; this draws a
// block per character so that layout is at least visible. The renderer package
// draws real text into an image and hands it to DrawImage instead.
func (m *monoFrameBuffer) DrawText(x, y int, text string, size int) error {
	charWidth := size / 2
	for i := range text {
		startX := x + i*charWidth
		if startX >= m.width {
			break
		}
		if err := m.DrawRect(startX, y, charWidth-1, size, false); err != nil {
			return err
		}
	}
	return nil
}

// DrawLine draws a horizontal line.
func (m *monoFrameBuffer) DrawLine(x, y, width int) error {
	for i := 0; i < width && x+i < m.width; i++ {
		if x+i >= 0 && y >= 0 && y < m.height {
			m.img.SetGray(x+i, y, color.Gray{Y: 255})
		}
	}
	return nil
}

// DrawPixel draws a single pixel.
func (m *monoFrameBuffer) DrawPixel(x, y int, on bool) error {
	if !m.inBounds(x, y) {
		return nil
	}

	if on {
		m.img.SetGray(x, y, color.Gray{Y: 255})
	} else {
		m.img.SetGray(x, y, color.Gray{Y: 0})
	}
	return nil
}

// DrawRect draws a rectangle, filled or outlined.
//
//nolint:gocyclo // drawing logic naturally has many conditional branches
func (m *monoFrameBuffer) DrawRect(x, y, width, height int, fill bool) error {
	if fill {
		for dy := 0; dy < height && y+dy < m.height; dy++ {
			for dx := 0; dx < width && x+dx < m.width; dx++ {
				if x+dx >= 0 && y+dy >= 0 {
					m.img.SetGray(x+dx, y+dy, color.Gray{Y: 255})
				}
			}
		}
	} else {
		for i := 0; i < width && x+i < m.width; i++ {
			if x+i >= 0 && y >= 0 {
				m.img.SetGray(x+i, y, color.Gray{Y: 255})
			}
			if x+i >= 0 && y+height-1 >= 0 && y+height-1 < m.height {
				m.img.SetGray(x+i, y+height-1, color.Gray{Y: 255})
			}
		}
		for i := 0; i < height && y+i < m.height; i++ {
			if x >= 0 && y+i >= 0 {
				m.img.SetGray(x, y+i, color.Gray{Y: 255})
			}
			if x+width-1 >= 0 && x+width-1 < m.width && y+i >= 0 {
				m.img.SetGray(x+width-1, y+i, color.Gray{Y: 255})
			}
		}
	}
	return nil
}

// DrawImage composites src onto the buffer at (x, y). The source is treated as
// monochrome: the brightest channel decides whether a pixel lights up, so
// saturated colours (e.g. pure green) render as white.
func (m *monoFrameBuffer) DrawImage(x, y int, img image.Image) error {
	bounds := img.Bounds()
	for dy := 0; dy < bounds.Dy() && y+dy < m.height; dy++ {
		for dx := 0; dx < bounds.Dx() && x+dx < m.width; dx++ {
			if x+dx < 0 || y+dy < 0 {
				continue
			}
			r, g, b, a := img.At(bounds.Min.X+dx, bounds.Min.Y+dy).RGBA()
			brightness := r
			if g > brightness {
				brightness = g
			}
			if b > brightness {
				brightness = b
			}
			on := brightness > luminanceThreshold && a > luminanceThreshold
			if on {
				m.img.SetGray(x+dx, y+dy, color.Gray{Y: 255})
			} else {
				m.img.SetGray(x+dx, y+dy, color.Gray{Y: 0})
			}
		}
	}
	return nil
}

// GetBounds returns the display dimensions.
func (m *monoFrameBuffer) GetBounds() image.Rectangle {
	return m.img.Bounds()
}

// GetBuffer returns the buffer packed into the page-major byte layout the
// SSD1306 and SH1106 GDDRAM use: one byte per (column, 8-row page), LSB = top.
func (m *monoFrameBuffer) GetBuffer() []byte {
	buf := make([]byte, m.width*m.height/8)
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			if m.lit(x, y) {
				byteIdx := x + (y/8)*m.width
				bitIdx := uint(y % 8) /* #nosec G115 -- modulo 8 is always 0–7 */
				buf[byteIdx] |= 1 << bitIdx
			}
		}
	}
	return buf
}

// lit reports whether the pixel at (x, y) should be lit. Out of bounds reads
// are treated as unlit so page packing can run past the panel edge.
func (m *monoFrameBuffer) lit(x, y int) bool {
	if !m.inBounds(x, y) {
		return false
	}
	return m.img.GrayAt(x, y).Y > 128
}

// pageBytes packs one 8-row page of the buffer into width bytes, LSB = top row.
func (m *monoFrameBuffer) pageBytes(page int) []byte {
	buf := make([]byte, m.width)
	for x := 0; x < m.width; x++ {
		var b byte
		for bit := range 8 {
			if m.lit(x, page*8+bit) {
				b |= 1 << bit
			}
		}
		buf[x] = b
	}
	return buf
}

// pages returns the number of 8-row GDDRAM pages the panel spans.
func (m *monoFrameBuffer) pages() int {
	return m.height / 8
}

// inBounds reports whether (x, y) addresses a pixel of the panel.
func (m *monoFrameBuffer) inBounds(x, y int) bool {
	return x >= 0 && x < m.width && y >= 0 && y < m.height
}
