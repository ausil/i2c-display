package renderer

import "image"

// Layout constants
const (
	// Margins
	MarginLeft  = 2
	MarginRight = 2
	MarginTop   = 2

	// Font sizes (approximations for basic font)
	FontHeight = 8
	FontWidth  = 6
)

// Legacy constants for backward compatibility (128x64 layout)
const (
	Line1Y = 0  // Hostname
	Line2Y = 10 // Separator
	Line3Y = 16 // First content line
	Line4Y = 28 // Second content line
	Line5Y = 40 // Third content line
	Line6Y = 52 // Fourth content line (status/page indicator)
)

// Layout holds adaptive layout information based on display size
type Layout struct {
	Width           int
	Height          int
	HeaderY         int   // Hostname position
	SeparatorY      int   // Separator line position
	ContentLines    []int // Y positions for content lines
	FooterY         int   // Footer/status line position
	ShowHeader      bool  // Whether to show hostname header
	ShowSeparator   bool  // Whether to show separator line
	MaxContentLines int   // Maximum content lines available
}

// NewLayout creates an adaptive layout based on display bounds
func NewLayout(bounds image.Rectangle) *Layout {
	width := bounds.Dx()
	height := bounds.Dy()

	layout := &Layout{
		Width:  width,
		Height: height,
	}

	// Adapt layout based on display height
	switch {
	case height <= 32:
		// Small display (128x32 or 96x16)
		// Only 32 pixels tall - compact layout
		// Font is 13px tall, so header (0-11) + one content line (14-25) fits cleanly
		layout.ShowHeader = true     // Always show hostname
		layout.ShowSeparator = false // Skip separator to save space
		layout.HeaderY = 0
		layout.SeparatorY = -1
		layout.ContentLines = []int{14} // Single content line with proper spacing
		layout.FooterY = -1             // No footer
		layout.MaxContentLines = 1

	case height <= 64:
		// Standard display (128x64)
		layout.ShowHeader = true
		layout.ShowSeparator = true
		layout.HeaderY = 0
		layout.SeparatorY = 10
		layout.ContentLines = []int{16, 28, 40} // 3 main content lines
		layout.FooterY = 52
		layout.MaxContentLines = 3

	default:
		// Large display (128x128 or bigger)
		layout.ShowHeader = true
		layout.ShowSeparator = true
		layout.HeaderY = 0
		layout.SeparatorY = 12
		// More content lines for larger displays
		layout.ContentLines = []int{20, 36, 52, 68, 84, 100}
		layout.FooterY = 116
		layout.MaxContentLines = 6
	}

	return layout
}

// CenterText calculates the X coordinate to center text
func CenterText(text string, displayWidth int) int {
	textWidth := len(text) * FontWidth
	return (displayWidth - textWidth) / 2
}
