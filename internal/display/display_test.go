package display

import (
	"fmt"
	"image"
	"image/color"
	"testing"
)

func TestMockDisplay(t *testing.T) {
	display := NewMockDisplay(128, 64)

	// Test Init
	if err := display.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	calls := display.GetCalls()
	if len(calls) != 1 || calls[0] != "Init" {
		t.Errorf("expected Init call, got %v", calls)
	}

	// Test Clear
	display.ClearCalls()
	if err := display.Clear(); err != nil {
		t.Fatalf("Clear() failed: %v", err)
	}

	calls = display.GetCalls()
	if len(calls) != 1 || calls[0] != "Clear" {
		t.Errorf("expected Clear call, got %v", calls)
	}

	// Test DrawText
	display.ClearCalls()
	if err := display.DrawText(0, 0, "Hello", FontSmall); err != nil {
		t.Fatalf("DrawText() failed: %v", err)
	}

	calls = display.GetCalls()
	if len(calls) != 1 {
		t.Errorf("expected 1 call, got %d: %v", len(calls), calls)
	}

	// Test DrawLine
	display.ClearCalls()
	if err := display.DrawLine(0, 10, 128); err != nil {
		t.Fatalf("DrawLine() failed: %v", err)
	}

	// Verify line was drawn
	for x := 0; x < 128; x++ {
		if !display.GetPixel(x, 10) {
			t.Errorf("expected pixel at (%d, 10) to be on", x)
		}
	}

	// Test DrawPixel
	display.Clear()
	if err := display.DrawPixel(50, 30, true); err != nil {
		t.Fatalf("DrawPixel() failed: %v", err)
	}

	if !display.GetPixel(50, 30) {
		t.Error("expected pixel at (50, 30) to be on")
	}

	// Test DrawRect
	display.Clear()
	if err := display.DrawRect(10, 10, 20, 15, false); err != nil {
		t.Fatalf("DrawRect() failed: %v", err)
	}

	// Check corners
	if !display.GetPixel(10, 10) {
		t.Error("expected top-left corner to be on")
	}
	if !display.GetPixel(29, 10) {
		t.Error("expected top-right corner to be on")
	}
	if !display.GetPixel(10, 24) {
		t.Error("expected bottom-left corner to be on")
	}
	if !display.GetPixel(29, 24) {
		t.Error("expected bottom-right corner to be on")
	}

	// Test filled rect
	display.Clear()
	if err := display.DrawRect(10, 10, 5, 5, true); err != nil {
		t.Fatalf("DrawRect(fill) failed: %v", err)
	}

	// Check all pixels in rect are on
	for y := 10; y < 15; y++ {
		for x := 10; x < 15; x++ {
			if !display.GetPixel(x, y) {
				t.Errorf("expected pixel at (%d, %d) to be on", x, y)
			}
		}
	}

	// Test Show
	display.ClearCalls()
	if err := display.Show(); err != nil {
		t.Fatalf("Show() failed: %v", err)
	}

	// Test Close
	display.ClearCalls()
	if err := display.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Test GetBounds
	bounds := display.GetBounds()
	if bounds.Dx() != 128 || bounds.Dy() != 64 {
		t.Errorf("expected bounds 128x64, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Test GetBuffer
	buffer := display.GetBuffer()
	expectedLen := 128 * 64 / 8
	if len(buffer) != expectedLen {
		t.Errorf("expected buffer length %d, got %d", expectedLen, len(buffer))
	}
}

func TestMockDisplayErrors(t *testing.T) {
	display := NewMockDisplay(128, 64)

	// Configure to return errors
	display.SetError(true, "test error")

	if err := display.Init(); err == nil {
		t.Error("expected Init to fail")
	}

	if err := display.Clear(); err == nil {
		t.Error("expected Clear to fail")
	}

	if err := display.DrawText(0, 0, "test", FontSmall); err == nil {
		t.Error("expected DrawText to fail")
	}

	if err := display.Show(); err == nil {
		t.Error("expected Show to fail")
	}
}

func TestMockDisplayDrawImage(t *testing.T) {
	display := NewMockDisplay(128, 64)

	// Create a small test image
	img := image.NewGray(image.Rect(0, 0, 10, 10))
	// Fill with white
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.SetGray(x, y, color.Gray{Y: 255})
		}
	}

	display.Clear()
	if err := display.DrawImage(0, 0, img); err != nil {
		t.Fatalf("DrawImage() failed: %v", err)
	}

	// Verify some pixels were set
	pixelCount := 0
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if display.GetPixel(x, y) {
				pixelCount++
			}
		}
	}

	if pixelCount == 0 {
		t.Error("expected some pixels to be set after DrawImage")
	}
}

func TestMockDisplayBoundaryConditions(t *testing.T) {
	display := NewMockDisplay(128, 64)

	// Test drawing outside bounds
	if err := display.DrawPixel(-1, -1, true); err != nil {
		t.Errorf("DrawPixel outside bounds should not error: %v", err)
	}

	if err := display.DrawPixel(200, 200, true); err != nil {
		t.Errorf("DrawPixel outside bounds should not error: %v", err)
	}

	// Test line partially outside bounds
	if err := display.DrawLine(100, 10, 100); err != nil {
		t.Fatalf("DrawLine() failed: %v", err)
	}

	// Test rect partially outside bounds
	if err := display.DrawRect(120, 60, 20, 20, false); err != nil {
		t.Fatalf("DrawRect() failed: %v", err)
	}
}

func TestMockDisplayConcurrency(t *testing.T) {
	display := NewMockDisplay(128, 64)

	// Test concurrent access
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			display.DrawPixel(10, 10, true)
			display.Clear()
			display.GetCalls()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestDisplayConstants(t *testing.T) {
	if FontSmall != 8 {
		t.Errorf("expected FontSmall=8, got %d", FontSmall)
	}
	if FontMedium != 12 {
		t.Errorf("expected FontMedium=12, got %d", FontMedium)
	}
	if FontLarge != 16 {
		t.Errorf("expected FontLarge=16, got %d", FontLarge)
	}
}

func TestMockDisplaySetBrightness(t *testing.T) {
	display := NewMockDisplay(128, 64)

	tests := []uint8{0, 127, 255, 64, 192}

	for _, level := range tests {
		if err := display.SetBrightness(level); err != nil {
			t.Errorf("SetBrightness(%d) failed: %v", level, err)
		}
	}
}

func TestMockDisplayGetPixelOutOfBounds(t *testing.T) {
	display := NewMockDisplay(128, 64)

	// Test that GetPixel returns false for out-of-bounds coordinates
	if display.GetPixel(-1, 0) {
		t.Error("expected false for negative x")
	}

	if display.GetPixel(0, -1) {
		t.Error("expected false for negative y")
	}

	if display.GetPixel(128, 0) {
		t.Error("expected false for x >= width")
	}

	if display.GetPixel(0, 64) {
		t.Error("expected false for y >= height")
	}
}

func TestMockDisplayDrawImagePartial(t *testing.T) {
	display := NewMockDisplay(128, 64)

	// Create image that extends beyond display bounds
	img := image.NewGray(image.Rect(0, 0, 50, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 50; x++ {
			img.SetGray(x, y, color.Gray{Y: 255})
		}
	}

	// Draw at position that causes partial clipping
	display.Clear()
	if err := display.DrawImage(100, 40, img); err != nil {
		t.Fatalf("DrawImage() failed: %v", err)
	}

	// Verify pixels within bounds are set
	if !display.GetPixel(100, 40) {
		t.Error("expected pixel at (100, 40) to be on")
	}

	// Verify we don't crash when drawing completely outside bounds
	display.Clear()
	if err := display.DrawImage(200, 200, img); err != nil {
		t.Fatalf("DrawImage() outside bounds failed: %v", err)
	}
}

func TestMockDisplayDrawImageGrayscale(t *testing.T) {
	display := NewMockDisplay(128, 64)

	// Create image with varying grayscale values
	img := image.NewGray(image.Rect(0, 0, 10, 10))

	// Test threshold behavior
	// Pixels with Y <= 128 should be off
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			img.SetGray(x, y, color.Gray{Y: 64}) // Below threshold
		}
	}

	// Pixels with Y > 128 should be on
	for y := 5; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.SetGray(x, y, color.Gray{Y: 255}) // Above threshold
		}
	}

	display.Clear()
	if err := display.DrawImage(0, 0, img); err != nil {
		t.Fatalf("DrawImage() failed: %v", err)
	}

	// Check that dark pixels are off
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			if display.GetPixel(x, y) {
				t.Errorf("expected pixel at (%d, %d) to be off", x, y)
			}
		}
	}

	// Check that bright pixels are on
	for y := 5; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if !display.GetPixel(x, y) {
				t.Errorf("expected pixel at (%d, %d) to be on", x, y)
			}
		}
	}
}

func TestMockDisplayDrawImageRGBA(t *testing.T) {
	display := NewMockDisplay(128, 64)

	// Create RGBA image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))

	// Set some pixels to white
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	display.Clear()
	if err := display.DrawImage(5, 5, img); err != nil {
		t.Fatalf("DrawImage() with RGBA failed: %v", err)
	}

	// Verify some pixels were set
	if !display.GetPixel(5, 5) {
		t.Error("expected pixel at (5, 5) to be on after RGBA image")
	}
}

func TestMockDisplayDifferentSizes(t *testing.T) {
	sizes := []struct {
		width  int
		height int
	}{
		{128, 64},
		{128, 32},
		{96, 16},
		{64, 48},
		{256, 64},
	}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("%dx%d", size.width, size.height), func(t *testing.T) {
			display := NewMockDisplay(size.width, size.height)

			bounds := display.GetBounds()
			if bounds.Dx() != size.width || bounds.Dy() != size.height {
				t.Errorf("expected bounds %dx%d, got %dx%d",
					size.width, size.height, bounds.Dx(), bounds.Dy())
			}

			expectedBufferLen := size.width * size.height / 8
			buffer := display.GetBuffer()
			if len(buffer) != expectedBufferLen {
				t.Errorf("expected buffer length %d, got %d",
					expectedBufferLen, len(buffer))
			}

			// Test that we can draw within bounds
			if err := display.DrawPixel(0, 0, true); err != nil {
				t.Errorf("DrawPixel failed: %v", err)
			}

			if err := display.DrawPixel(size.width-1, size.height-1, true); err != nil {
				t.Errorf("DrawPixel at max bounds failed: %v", err)
			}
		})
	}
}

func TestMockDisplayBufferContent(t *testing.T) {
	display := NewMockDisplay(128, 32)
	display.Clear()

	// Set specific pixels and verify buffer
	display.DrawPixel(0, 0, true)
	display.DrawPixel(127, 31, true)
	display.DrawPixel(64, 16, true)

	buffer := display.GetBuffer()

	// Verify buffer is not all zeros
	allZero := true
	for _, b := range buffer {
		if b != 0 {
			allZero = false
			break
		}
	}

	if allZero {
		t.Error("expected buffer to have some non-zero bytes after drawing pixels")
	}
}

func TestMockDisplayLineEdgeCases(t *testing.T) {
	display := NewMockDisplay(128, 64)

	tests := []struct {
		name  string
		x     int
		y     int
		width int
	}{
		{"zero width", 0, 0, 0},
		{"single pixel", 0, 0, 1},
		{"full width", 0, 32, 128},
		{"partial beyond bounds", 120, 10, 20},
		{"negative x", -10, 10, 20},
		{"negative y", 10, -1, 10},
		{"beyond height", 0, 70, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			display.Clear()
			if err := display.DrawLine(tt.x, tt.y, tt.width); err != nil {
				t.Errorf("DrawLine(%d, %d, %d) failed: %v", tt.x, tt.y, tt.width, err)
			}
		})
	}
}

func TestMockDisplayRectEdgeCases(t *testing.T) {
	display := NewMockDisplay(128, 64)

	tests := []struct {
		name   string
		x      int
		y      int
		width  int
		height int
		fill   bool
	}{
		{"zero size", 0, 0, 0, 0, false},
		{"single pixel filled", 10, 10, 1, 1, true},
		{"single pixel outline", 10, 10, 1, 1, false},
		{"negative x filled", -5, 10, 10, 10, true},
		{"negative y outline", 10, -5, 10, 10, false},
		{"beyond bounds filled", 120, 60, 20, 20, true},
		{"beyond bounds outline", 120, 60, 20, 20, false},
		{"tall rect", 50, 0, 10, 64, false},
		{"wide rect", 0, 30, 128, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			display.Clear()
			if err := display.DrawRect(tt.x, tt.y, tt.width, tt.height, tt.fill); err != nil {
				t.Errorf("DrawRect(%d, %d, %d, %d, %v) failed: %v",
					tt.x, tt.y, tt.width, tt.height, tt.fill, err)
			}
		})
	}
}

func TestMockDisplayTextSizes(t *testing.T) {
	display := NewMockDisplay(128, 64)

	texts := []struct {
		text string
		size int
	}{
		{"A", FontSmall},
		{"Hello", FontMedium},
		{"Long text that exceeds display width", FontLarge},
		{"", FontSmall},
		{"Test", 20},
	}

	for _, tt := range texts {
		t.Run(fmt.Sprintf("%s-%d", tt.text, tt.size), func(t *testing.T) {
			display.Clear()
			if err := display.DrawText(0, 0, tt.text, tt.size); err != nil {
				t.Errorf("DrawText failed: %v", err)
			}
		})
	}
}

func TestMockDisplayErrorRecovery(t *testing.T) {
	display := NewMockDisplay(128, 64)

	// Set error state
	display.SetError(true, "test error")

	// Verify operations fail
	if err := display.Init(); err == nil {
		t.Error("expected Init to fail in error state")
	}

	// Clear error state
	display.SetError(false, "")

	// Verify operations succeed again
	if err := display.Init(); err != nil {
		t.Errorf("expected Init to succeed after clearing error: %v", err)
	}

	if err := display.Clear(); err != nil {
		t.Errorf("expected Clear to succeed after clearing error: %v", err)
	}
}

func TestMockDisplayCallTracking(t *testing.T) {
	display := NewMockDisplay(128, 64)

	// Perform sequence of operations
	display.Init()
	display.Clear()
	display.DrawPixel(0, 0, true)
	display.DrawLine(0, 10, 128)
	display.Show()

	calls := display.GetCalls()
	if len(calls) != 5 {
		t.Errorf("expected 5 calls, got %d: %v", len(calls), calls)
	}

	// Check that expected call types are present (calls may include parameters)
	expectedPrefixes := []string{"Init", "Clear", "DrawPixel", "DrawLine", "Show"}
	for i, expectedPrefix := range expectedPrefixes {
		if i >= len(calls) {
			t.Errorf("missing call %d: expected prefix %s", i, expectedPrefix)
			continue
		}
		// Check if call starts with expected prefix
		if len(calls[i]) < len(expectedPrefix) || calls[i][:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("call %d: expected to start with %s, got %s", i, expectedPrefix, calls[i])
		}
	}

	// Test ClearCalls
	display.ClearCalls()
	calls = display.GetCalls()
	if len(calls) != 0 {
		t.Errorf("expected 0 calls after ClearCalls, got %d", len(calls))
	}
}
