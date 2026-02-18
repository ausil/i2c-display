package renderer

import (
	"image"
	"testing"

	"github.com/ausil/i2c-display/internal/display"
)

func TestIconDimensions(t *testing.T) {
	initIcons()

	icons := map[string]*iconInfo{
		"disk":   {iconDisk},
		"memory": {iconMemory},
		"cpu":    {iconCPU},
	}

	for name, ic := range icons {
		t.Run(name, func(t *testing.T) {
			bounds := ic.img.Bounds()
			if bounds.Dx() != IconWidth {
				t.Errorf("width = %d, want %d", bounds.Dx(), IconWidth)
			}
			if bounds.Dy() != IconHeight {
				t.Errorf("height = %d, want %d", bounds.Dy(), IconHeight)
			}
		})
	}
}

func TestIconsNotBlank(t *testing.T) {
	initIcons()

	icons := map[string]*iconInfo{
		"disk":   {iconDisk},
		"memory": {iconMemory},
		"cpu":    {iconCPU},
	}

	for name, ic := range icons {
		t.Run(name, func(t *testing.T) {
			nonZero := 0
			bounds := ic.img.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					if ic.img.GrayAt(x, y).Y > 0 {
						nonZero++
					}
				}
			}
			if nonZero == 0 {
				t.Errorf("icon %q has no non-zero pixels", name)
			}
		})
	}
}

func TestDrawIconText(t *testing.T) {
	initIcons()
	disp := display.NewMockDisplay(128, 64)

	if err := DrawIconText(disp, 1, 16, iconDisk, "50.0%"); err != nil {
		t.Fatalf("DrawIconText failed: %v", err)
	}

	calls := disp.GetCalls()
	if len(calls) < 2 {
		t.Errorf("expected at least 2 draw calls (image + text), got %d", len(calls))
	}
}

// iconInfo is a small wrapper so we can range over icons in table tests.
type iconInfo struct {
	img *image.Gray
}
