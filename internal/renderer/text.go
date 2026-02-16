package renderer

import (
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"github.com/ausil/i2c-display/internal/display"
)

// Colours used for rendering on colour displays.
// On monochrome displays these are thresholded to white.
var (
	ColorGreen = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
)

// DrawText renders text at the specified position using a simple bitmap font
func DrawText(disp display.Display, x, y int, text string) error {
	// Use basicfont (7x13 font)
	face := basicfont.Face7x13

	// Measure text to create appropriately sized image
	width := font.MeasureString(face, text).Ceil()
	height := int(face.Metrics().Ascent.Ceil()) + int(face.Metrics().Descent.Ceil())

	// Create an image just large enough for the text
	textImg := image.NewGray(image.Rect(0, 0, width, height))

	// Create drawer with origin at (0, ascent) in the small image
	drawer := &font.Drawer{
		Dst:  textImg,
		Src:  image.White,
		Face: face,
		Dot:  fixed.P(0, int(face.Metrics().Ascent.Ceil())),
	}

	// Draw the text
	drawer.DrawString(text)

	// Draw the text image at the specified position
	return disp.DrawImage(x, y, textImg)
}

// DrawTextCentered draws text centered horizontally
func DrawTextCentered(disp display.Display, y int, text string) error {
	bounds := disp.GetBounds()
	face := basicfont.Face7x13
	width := font.MeasureString(face, text).Ceil()
	x := (bounds.Dx() - width) / 2
	return DrawText(disp, x, y, text)
}

// DrawTextColor renders text in a specific colour at the given position.
// On colour displays the colour is preserved; on monochrome displays
// any bright colour is rendered as white.
func DrawTextColor(disp display.Display, x, y int, text string, c color.Color) error {
	face := basicfont.Face7x13
	width := font.MeasureString(face, text).Ceil()
	height := int(face.Metrics().Ascent.Ceil()) + int(face.Metrics().Descent.Ceil())

	textImg := image.NewNRGBA(image.Rect(0, 0, width, height))
	drawer := &font.Drawer{
		Dst:  textImg,
		Src:  &image.Uniform{c},
		Face: face,
		Dot:  fixed.P(0, int(face.Metrics().Ascent.Ceil())),
	}
	drawer.DrawString(text)
	return disp.DrawImage(x, y, textImg)
}

// DrawTextCenteredColor draws coloured text centered horizontally.
func DrawTextCenteredColor(disp display.Display, y int, text string, c color.Color) error {
	bounds := disp.GetBounds()
	face := basicfont.Face7x13
	width := font.MeasureString(face, text).Ceil()
	x := (bounds.Dx() - width) / 2
	return DrawTextColor(disp, x, y, text, c)
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
