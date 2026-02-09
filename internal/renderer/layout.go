package renderer

// Layout constants for 128x64 display
const (
	// Display dimensions
	Width  = 128
	Height = 64

	// Line positions (Y coordinates)
	Line1Y = 0  // Hostname
	Line2Y = 10 // Separator
	Line3Y = 16 // First content line
	Line4Y = 28 // Second content line
	Line5Y = 40 // Third content line
	Line6Y = 52 // Fourth content line (status/page indicator)

	// Margins
	MarginLeft  = 2
	MarginRight = 2
	MarginTop   = 2

	// Font sizes (approximations for basic font)
	FontHeight = 8
	FontWidth  = 6
)

// CenterText calculates the X coordinate to center text
func CenterText(text string, displayWidth int) int {
	textWidth := len(text) * FontWidth
	return (displayWidth - textWidth) / 2
}
