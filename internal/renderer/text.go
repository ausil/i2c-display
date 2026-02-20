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
	ColorGreen  = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	ColorYellow = color.NRGBA{R: 255, G: 255, B: 0, A: 255}
	ColorRed    = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
)

// MetricColor returns green/yellow/red based on a usage percentage.
// 0-60% → green, 60-85% → yellow, >85% → red.
func MetricColor(percent float64) color.NRGBA {
	switch {
	case percent > 85:
		return ColorRed
	case percent >= 60:
		return ColorYellow
	default:
		return ColorGreen
	}
}

// TempColor returns green/yellow/red based on CPU temperature in Celsius.
// <55C → green, 55-75C → yellow, >75C → red.
func TempColor(celsius float64) color.NRGBA {
	switch {
	case celsius > 75:
		return ColorRed
	case celsius >= 55:
		return ColorYellow
	default:
		return ColorGreen
	}
}

// LoadColor returns green/yellow/red based on load average per CPU core.
// loadAvg/numCPU < 0.7 → green, 0.7–1.0 → yellow, > 1.0 → red.
func LoadColor(loadAvg float64, numCPU int) color.NRGBA {
	if numCPU <= 0 {
		numCPU = 1
	}
	perCore := loadAvg / float64(numCPU)
	switch {
	case perCore > 1.0:
		return ColorRed
	case perCore >= 0.7:
		return ColorYellow
	default:
		return ColorGreen
	}
}

// DrawText renders text at the specified position using a simple bitmap font
func DrawText(disp display.Display, x, y int, text string) error {
	// Use basicfont (7x13 font)
	face := basicfont.Face7x13

	// Measure text to create appropriately sized image
	width := font.MeasureString(face, text).Ceil()
	height := face.Metrics().Ascent.Ceil() + face.Metrics().Descent.Ceil()

	// Create an image just large enough for the text
	textImg := image.NewGray(image.Rect(0, 0, width, height))

	// Create drawer with origin at (0, ascent) in the small image
	drawer := &font.Drawer{
		Dst:  textImg,
		Src:  image.White,
		Face: face,
		Dot:  fixed.P(0, face.Metrics().Ascent.Ceil()),
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
	height := face.Metrics().Ascent.Ceil() + face.Metrics().Descent.Ceil()

	textImg := image.NewNRGBA(image.Rect(0, 0, width, height))
	drawer := &font.Drawer{
		Dst:  textImg,
		Src:  &image.Uniform{c},
		Face: face,
		Dot:  fixed.P(0, face.Metrics().Ascent.Ceil()),
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

// ScaledTextHeight returns the rendered pixel height of the font used for the
// given scale factor. A scale of 0 or 1 uses the full-size basicfont (13 px);
// any other value in (0,1) uses Face5x7 (7 px).
func ScaledTextHeight(scale float64) int {
	if scale > 0 && scale < 1 {
		return font5x7GlyphHeight // 7 px — matches Face5x7
	}
	return 13 // basicfont.Face7x13
}

// DrawTextColorScaled renders text in colour using the appropriate font for the
// given scale factor. scale=0 or scale=1 uses the standard 7×13 basicfont;
// any value in (0,1) uses the compact 5×7 font (Face5x7) directly, which is
// far more legible than downsampling the larger font.
func DrawTextColorScaled(disp display.Display, x, y int, text string, c color.Color, scale float64) error {
	var face font.Face
	if scale > 0 && scale < 1 {
		face = Face5x7
	} else {
		face = basicfont.Face7x13
	}
	width := font.MeasureString(face, text).Ceil()
	height := face.Metrics().Ascent.Ceil() + face.Metrics().Descent.Ceil()

	textImg := image.NewNRGBA(image.Rect(0, 0, width, height))
	drawer := &font.Drawer{
		Dst:  textImg,
		Src:  &image.Uniform{c},
		Face: face,
		Dot:  fixed.P(0, face.Metrics().Ascent.Ceil()),
	}
	drawer.DrawString(text)
	return disp.DrawImage(x, y, textImg)
}

// DrawTextCenteredColorScaled draws centred coloured text using the font
// appropriate for the given scale factor (see DrawTextColorScaled).
func DrawTextCenteredColorScaled(disp display.Display, y int, text string, c color.Color, scale float64) error {
	bounds := disp.GetBounds()
	var face font.Face
	if scale > 0 && scale < 1 {
		face = Face5x7
	} else {
		face = basicfont.Face7x13
	}
	width := font.MeasureString(face, text).Ceil()
	x := (bounds.Dx() - width) / 2
	return DrawTextColorScaled(disp, x, y, text, c, scale)
}

// MeasureTextSmall returns the pixel width of text rendered with Face5x7.
func MeasureTextSmall(text string) int {
	return font.MeasureString(Face5x7, text).Ceil()
}

// TruncateTextSmall truncates text to fit within maxWidth pixels as measured
// by Face5x7, appending "..." when truncation occurs.
func TruncateTextSmall(text string, maxWidth int) string {
	if MeasureTextSmall(text) <= maxWidth {
		return text
	}

	ellipsis := "..."
	ellipsisWidth := MeasureTextSmall(ellipsis)
	availableWidth := maxWidth - ellipsisWidth

	left, right := 0, len(text)
	result := text

	for left < right {
		mid := (left + right + 1) / 2
		if MeasureTextSmall(text[:mid]) <= availableWidth {
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
