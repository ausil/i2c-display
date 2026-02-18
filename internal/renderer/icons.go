package renderer

import (
	"image"
	"image/color"
	"sync"

	"github.com/ausil/i2c-display/internal/display"
)

// Icon layout constants.
const (
	IconWidth  = 10
	IconHeight = 11
	IconGap    = 2 // pixels between icon and text
)

// Icon bitmap data (0 = off, 255 = on).
// Each icon is an [IconHeight][IconWidth]byte array.

// diskBitmap is a cylinder/barrel shape representing storage.
//
//	__XXXXXX__
//	_XXXXXXXX_
//	XX______XX
//	XX______XX
//	XX______XX
//	_XXXXXXXX_
//	XX______XX
//	XX______XX
//	XX______XX
//	_XXXXXXXX_
//	__XXXXXX__
//
//nolint:dupl // bitmap pixel data — not a logic duplicate
var diskBitmap = [IconHeight][IconWidth]byte{
	{0, 0, 255, 255, 255, 255, 255, 255, 0, 0},
	{0, 255, 255, 255, 255, 255, 255, 255, 255, 0},
	{255, 255, 0, 0, 0, 0, 0, 0, 255, 255},
	{255, 255, 0, 0, 0, 0, 0, 0, 255, 255},
	{255, 255, 0, 0, 0, 0, 0, 0, 255, 255},
	{0, 255, 255, 255, 255, 255, 255, 255, 255, 0},
	{255, 255, 0, 0, 0, 0, 0, 0, 255, 255},
	{255, 255, 0, 0, 0, 0, 0, 0, 255, 255},
	{255, 255, 0, 0, 0, 0, 0, 0, 255, 255},
	{0, 255, 255, 255, 255, 255, 255, 255, 255, 0},
	{0, 0, 255, 255, 255, 255, 255, 255, 0, 0},
}

// memoryBitmap is a chip/DIMM with pins at the bottom.
//
//	_XXXXXXXX_
//	_X______X_
//	_X_XXXX_X_
//	_X_XXXX_X_
//	_X______X_
//	_X_XXXX_X_
//	_X_XXXX_X_
//	_X______X_
//	_XXXXXXXX_
//	X_X_XX_X_X
//	X_X_XX_X_X
//
//nolint:dupl // bitmap pixel data — not a logic duplicate
var memoryBitmap = [IconHeight][IconWidth]byte{
	{0, 255, 255, 255, 255, 255, 255, 255, 255, 0},
	{0, 255, 0, 0, 0, 0, 0, 0, 255, 0},
	{0, 255, 0, 255, 255, 255, 255, 0, 255, 0},
	{0, 255, 0, 255, 255, 255, 255, 0, 255, 0},
	{0, 255, 0, 0, 0, 0, 0, 0, 255, 0},
	{0, 255, 0, 255, 255, 255, 255, 0, 255, 0},
	{0, 255, 0, 255, 255, 255, 255, 0, 255, 0},
	{0, 255, 0, 0, 0, 0, 0, 0, 255, 0},
	{0, 255, 255, 255, 255, 255, 255, 255, 255, 0},
	{255, 0, 255, 0, 255, 255, 0, 255, 0, 255},
	{255, 0, 255, 0, 255, 255, 0, 255, 0, 255},
}

// cpuTempBitmap is a thermometer shape (CPU shows temperature).
//
//	___XXXX___
//	__X____X__
//	__X_XX_X__
//	__X____X__
//	__X_XX_X__
//	__X____X__
//	__X_XX_X__
//	_XX_XX_XX_
//	_XXXXXXXX_
//	_XXXXXXXX_
//	__XXXXXX__
//
//nolint:dupl // bitmap pixel data — not a logic duplicate
var cpuTempBitmap = [IconHeight][IconWidth]byte{
	{0, 0, 0, 255, 255, 255, 255, 0, 0, 0},
	{0, 0, 255, 0, 0, 0, 0, 255, 0, 0},
	{0, 0, 255, 0, 255, 255, 0, 255, 0, 0},
	{0, 0, 255, 0, 0, 0, 0, 255, 0, 0},
	{0, 0, 255, 0, 255, 255, 0, 255, 0, 0},
	{0, 0, 255, 0, 0, 0, 0, 255, 0, 0},
	{0, 0, 255, 0, 255, 255, 0, 255, 0, 0},
	{0, 255, 255, 0, 255, 255, 0, 255, 255, 0},
	{0, 255, 255, 255, 255, 255, 255, 255, 255, 0},
	{0, 255, 255, 255, 255, 255, 255, 255, 255, 0},
	{0, 0, 255, 255, 255, 255, 255, 255, 0, 0},
}

// Cached icon images, initialized lazily via sync.Once.
var (
	iconDisk   *image.Gray
	iconMemory *image.Gray
	iconCPU    *image.Gray
	iconsOnce  sync.Once
)

// initIcons lazily creates the Gray images from bitmap data.
func initIcons() {
	iconsOnce.Do(func() {
		iconDisk = bitmapToGray(&diskBitmap)
		iconMemory = bitmapToGray(&memoryBitmap)
		iconCPU = bitmapToGray(&cpuTempBitmap)
	})
}

// bitmapToGray converts a fixed-size bitmap array to an *image.Gray.
func bitmapToGray(bitmap *[IconHeight][IconWidth]byte) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, IconWidth, IconHeight))
	for y := 0; y < IconHeight; y++ {
		for x := 0; x < IconWidth; x++ {
			img.SetGray(x, y, color.Gray{Y: bitmap[y][x]})
		}
	}
	return img
}

// DrawIconText draws an icon followed by text on the same line.
// The icon is drawn at (x, y+1) to vertically center it within the
// 13-pixel text line height, then the text is drawn at
// (x + IconWidth + IconGap, y).
func DrawIconText(disp display.Display, x, y int, icon *image.Gray, text string) error {
	if err := disp.DrawImage(x, y+1, icon); err != nil {
		return err
	}
	return DrawText(disp, x+IconWidth+IconGap, y, text)
}

// DrawIconTextColor draws an icon (white) followed by coloured text.
// The icon is drawn at (x, y+1), then the text is drawn at
// (x + IconWidth + IconGap, y) in colour c.
func DrawIconTextColor(disp display.Display, x, y int, icon *image.Gray, text string, c color.Color) error {
	if err := disp.DrawImage(x, y+1, icon); err != nil {
		return err
	}
	return DrawTextColor(disp, x+IconWidth+IconGap, y, text, c)
}
