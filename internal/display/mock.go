package display

import (
	"fmt"
	"image"
	"strings"
	"sync"
)

// MockDisplay is a mock implementation for testing
type MockDisplay struct {
	mu          sync.Mutex
	initialized bool
	width       int
	height      int
	buffer      []byte
	calls       []string
	shouldError bool
	errorMsg    string
}

// NewMockDisplay creates a new mock display
func NewMockDisplay(width, height int) *MockDisplay {
	return &MockDisplay{
		width:  width,
		height: height,
		buffer: make([]byte, width*height/8),
		calls:  make([]string, 0),
	}
}

// SetError configures the mock to return an error
func (m *MockDisplay) SetError(shouldError bool, msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = shouldError
	m.errorMsg = msg
}

// GetCalls returns all recorded method calls
func (m *MockDisplay) GetCalls() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.calls...)
}

// ClearCalls clears the recorded calls
func (m *MockDisplay) ClearCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = make([]string, 0)
}

func (m *MockDisplay) recordCall(method string, args ...interface{}) {
	call := method
	if len(args) > 0 {
		call += fmt.Sprintf("(%v)", args)
	}
	m.calls = append(m.calls, call)
}

func (m *MockDisplay) checkError() error {
	if m.shouldError {
		return fmt.Errorf("%s", m.errorMsg)
	}
	return nil
}

// Init initializes the mock display
func (m *MockDisplay) Init() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("Init")

	if err := m.checkError(); err != nil {
		return err
	}

	m.initialized = true
	return nil
}

// Clear clears the display buffer
func (m *MockDisplay) Clear() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("Clear")

	if err := m.checkError(); err != nil {
		return err
	}

	for i := range m.buffer {
		m.buffer[i] = 0
	}
	return nil
}

// DrawText simulates drawing text
func (m *MockDisplay) DrawText(x, y int, text string, size int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("DrawText", x, y, text, size)

	if err := m.checkError(); err != nil {
		return err
	}

	// Simulate text rendering by setting pixels in approximate area
	charWidth := size / 2
	for i, ch := range text {
		startX := x + i*charWidth
		if startX >= m.width {
			break
		}
		// Set a few pixels to simulate the character
		for dx := 0; dx < charWidth && startX+dx < m.width; dx++ {
			for dy := 0; dy < size && y+dy < m.height; dy++ {
				m.setPixel(startX+dx, y+dy, true)
			}
		}
		_ = ch // Avoid unused variable
	}

	return nil
}

// DrawLine draws a horizontal line
func (m *MockDisplay) DrawLine(x, y, width int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("DrawLine", x, y, width)

	if err := m.checkError(); err != nil {
		return err
	}

	for i := 0; i < width && x+i < m.width; i++ {
		m.setPixel(x+i, y, true)
	}
	return nil
}

// DrawPixel draws a single pixel
func (m *MockDisplay) DrawPixel(x, y int, on bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("DrawPixel", x, y, on)

	if err := m.checkError(); err != nil {
		return err
	}

	m.setPixel(x, y, on)
	return nil
}

// DrawRect draws a rectangle
func (m *MockDisplay) DrawRect(x, y, width, height int, fill bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("DrawRect", x, y, width, height, fill)

	if err := m.checkError(); err != nil {
		return err
	}

	if fill {
		for dy := 0; dy < height && y+dy < m.height; dy++ {
			for dx := 0; dx < width && x+dx < m.width; dx++ {
				m.setPixel(x+dx, y+dy, true)
			}
		}
	} else {
		// Draw outline
		for i := 0; i < width && x+i < m.width; i++ {
			m.setPixel(x+i, y, true)
			if y+height-1 < m.height {
				m.setPixel(x+i, y+height-1, true)
			}
		}
		for i := 0; i < height && y+i < m.height; i++ {
			m.setPixel(x, y+i, true)
			if x+width-1 < m.width {
				m.setPixel(x+width-1, y+i, true)
			}
		}
	}
	return nil
}

// DrawImage draws an image
func (m *MockDisplay) DrawImage(x, y int, img image.Image) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("DrawImage", x, y, img.Bounds())

	if err := m.checkError(); err != nil {
		return err
	}

	bounds := img.Bounds()
	for dy := 0; dy < bounds.Dy() && y+dy < m.height; dy++ {
		for dx := 0; dx < bounds.Dx() && x+dx < m.width; dx++ {
			r, g, b, a := img.At(bounds.Min.X+dx, bounds.Min.Y+dy).RGBA()
			// Simple threshold: if pixel is bright enough and not transparent, turn on
			brightness := (r + g + b) / 3
			on := brightness > 32768 && a > 32768
			m.setPixel(x+dx, y+dy, on)
		}
	}
	return nil
}

// Show simulates flushing to hardware
func (m *MockDisplay) Show() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("Show")

	return m.checkError()
}

// Close simulates closing the display
func (m *MockDisplay) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("Close")

	if err := m.checkError(); err != nil {
		return err
	}

	m.initialized = false
	return nil
}

// GetBounds returns the display dimensions
func (m *MockDisplay) GetBounds() image.Rectangle {
	m.mu.Lock()
	defer m.mu.Unlock()
	return image.Rect(0, 0, m.width, m.height)
}

// GetBuffer returns a copy of the display buffer
func (m *MockDisplay) GetBuffer() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	buf := make([]byte, len(m.buffer))
	copy(buf, m.buffer)
	return buf
}

// GetPixel returns the state of a pixel (for testing)
func (m *MockDisplay) GetPixel(x, y int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.getPixel(x, y)
}

// setPixel sets a pixel (must be called with lock held)
func (m *MockDisplay) setPixel(x, y int, on bool) {
	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return
	}

	byteIdx := x + (y/8)*m.width
	bitIdx := uint(y % 8) /* #nosec G115 -- modulo 8 is always 0–7 */

	if byteIdx >= len(m.buffer) {
		return
	}

	if on {
		m.buffer[byteIdx] |= 1 << bitIdx
	} else {
		m.buffer[byteIdx] &^= 1 << bitIdx
	}
}

// getPixel gets a pixel state (must be called with lock held)
func (m *MockDisplay) getPixel(x, y int) bool {
	if x < 0 || x >= m.width || y < 0 || y >= m.height {
		return false
	}

	byteIdx := x + (y/8)*m.width
	bitIdx := uint(y % 8) /* #nosec G115 -- modulo 8 is always 0–7 */

	if byteIdx >= len(m.buffer) {
		return false
	}

	return (m.buffer[byteIdx] & (1 << bitIdx)) != 0
}

// String returns a simple ASCII representation of the display
func (m *MockDisplay) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var sb strings.Builder
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			if m.getPixel(x, y) {
				sb.WriteString("█")
			} else {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// SetBrightness simulates setting display brightness
func (m *MockDisplay) SetBrightness(level uint8) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordCall("SetBrightness", level)

	// Mock just records the call, no actual brightness control
	return m.checkError()
}
