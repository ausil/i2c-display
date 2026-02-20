package display

import (
	"image"
	"image/color"
)

// drawRectNRGBA draws a white rectangle (outline or filled) into an NRGBA image buffer.
//
//nolint:gocyclo // drawing logic naturally has many conditional branches
func drawRectNRGBA(img *image.NRGBA, x, y, width, height, imgWidth, imgHeight int, fill bool) {
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	if fill {
		for dy := 0; dy < height && y+dy < imgHeight; dy++ {
			for dx := 0; dx < width && x+dx < imgWidth; dx++ {
				if x+dx >= 0 && y+dy >= 0 {
					img.SetNRGBA(x+dx, y+dy, white)
				}
			}
		}
	} else {
		for i := 0; i < width && x+i < imgWidth; i++ {
			if x+i >= 0 && y >= 0 {
				img.SetNRGBA(x+i, y, white)
			}
			if x+i >= 0 && y+height-1 >= 0 && y+height-1 < imgHeight {
				img.SetNRGBA(x+i, y+height-1, white)
			}
		}
		for i := 0; i < height && y+i < imgHeight; i++ {
			if x >= 0 && y+i >= 0 {
				img.SetNRGBA(x, y+i, white)
			}
			if x+width-1 >= 0 && x+width-1 < imgWidth && y+i >= 0 {
				img.SetNRGBA(x+width-1, y+i, white)
			}
		}
	}
}

// drawImageNRGBA composites a source image into an NRGBA buffer at position (x, y).
func drawImageNRGBA(img *image.NRGBA, x, y, imgWidth, imgHeight int, src image.Image) {
	bounds := src.Bounds()
	for dy := 0; dy < bounds.Dy() && y+dy < imgHeight; dy++ {
		for dx := 0; dx < bounds.Dx() && x+dx < imgWidth; dx++ {
			if x+dx < 0 || y+dy < 0 {
				continue
			}
			r, g, b, a := src.At(bounds.Min.X+dx, bounds.Min.Y+dy).RGBA()
			if a > 32768 {
				img.SetNRGBA(x+dx, y+dy, color.NRGBA{
					R: uint8(r >> 8), /* #nosec G115 -- RGBA() >> 8 always fits uint8 */
					G: uint8(g >> 8), /* #nosec G115 -- RGBA() >> 8 always fits uint8 */
					B: uint8(b >> 8), /* #nosec G115 -- RGBA() >> 8 always fits uint8 */
					A: 255,
				})
			} else {
				img.SetNRGBA(x+dx, y+dy, color.NRGBA{A: 255})
			}
		}
	}
}
