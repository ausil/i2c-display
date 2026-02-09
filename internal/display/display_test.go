package display

import (
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
