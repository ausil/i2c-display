package renderer

import (
	"image"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/denniskorablev/ssd1306-display/internal/display"
)

// DrawText renders text at the specified position using a simple bitmap font
func DrawText(disp display.Display, x, y int, text string) error {
	// Create a small image for the text
	bounds := disp.GetBounds()
	textImg := image.NewGray(bounds)

	// Use basicfont (7x13 font)
	face := basicfont.Face7x13

	// Create drawer
	drawer := &font.Drawer{
		Dst:  textImg,
		Src:  image.White,
		Face: face,
		Dot:  fixed.P(x, y+int(face.Metrics().Ascent.Ceil())),
	}

	// Draw the text
	drawer.DrawString(text)

	// Draw the text image to the display
	return disp.DrawImage(0, 0, textImg)
}

// DrawTextCentered draws text centered horizontally
func DrawTextCentered(disp display.Display, y int, text string) error {
	bounds := disp.GetBounds()
	face := basicfont.Face7x13
	width := font.MeasureString(face, text).Ceil()
	x := (bounds.Dx() - width) / 2
	return DrawText(disp, x, y, text)
}

// DrawLine draws a horizontal line (used for separator)
func DrawLine(disp display.Display, y int) error {
	bounds := disp.GetBounds()
	return disp.DrawLine(MarginLeft, y, bounds.Dx()-MarginLeft-MarginRight)
}

// MeasureText returns the width of text in pixels
func MeasureText(text string) int {
	face := basicfont.Face7x13
	return font.MeasureString(face, text).Ceil()
}

// TruncateText truncates text to fit within maxWidth, adding "..." if needed
func TruncateText(text string, maxWidth int) string {
	if MeasureText(text) <= maxWidth {
		return text
	}

	// Binary search for the right length
	ellipsis := "..."
	ellipsisWidth := MeasureText(ellipsis)
	availableWidth := maxWidth - ellipsisWidth

	left, right := 0, len(text)
	result := text

	for left < right {
		mid := (left + right + 1) / 2
		if MeasureText(text[:mid]) <= availableWidth {
			result = text[:mid]
			left = mid
		} else {
			right = mid - 1
		}
	}

	if len(result) < len(text) {
		return result + ellipsis
	}

	return text
}
