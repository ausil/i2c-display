package screensaver

import (
	"context"
	"testing"
	"time"

	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/logger"
)

func TestNew(t *testing.T) {
	cfg := Config{
		Enabled:          true,
		Mode:             ModeDim,
		IdleTimeout:      30 * time.Second,
		DimBrightness:    50,
		NormalBrightness: 255,
	}

	disp := display.NewMockDisplay(128, 64)
	log := logger.NewDefault()

	ss := New(cfg, disp, log)
	if ss == nil {
		t.Fatal("expected screen saver, got nil")
	}

	if ss.IsActive() {
		t.Error("screen saver should not be active initially")
	}
}

func TestScreenSaverDisabled(t *testing.T) {
	cfg := Config{
		Enabled: false,
	}

	disp := display.NewMockDisplay(128, 64)
	log := logger.NewDefault()

	ss := New(cfg, disp, log)
	ctx := context.Background()

	if err := ss.Start(ctx); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// ResetActivity should be no-op when disabled
	ss.ResetActivity()

	ss.Stop()
}

func TestScreenSaverActivation(t *testing.T) {
	cfg := Config{
		Enabled:          true,
		Mode:             ModeDim,
		IdleTimeout:      100 * time.Millisecond,
		DimBrightness:    50,
		NormalBrightness: 255,
	}

	disp := display.NewMockDisplay(128, 64)
	log := logger.NewDefault()

	ss := New(cfg, disp, log)

	// Manually trigger check after timeout
	time.Sleep(150 * time.Millisecond)
	ss.check()

	if !ss.IsActive() {
		t.Error("screen saver should be active after timeout")
	}
}

func TestScreenSaverDeactivation(t *testing.T) {
	cfg := Config{
		Enabled:          true,
		Mode:             ModeDim,
		IdleTimeout:      50 * time.Millisecond,
		DimBrightness:    50,
		NormalBrightness: 255,
	}

	disp := display.NewMockDisplay(128, 64)
	log := logger.NewDefault()

	ss := New(cfg, disp, log)

	// Wait for activation
	time.Sleep(100 * time.Millisecond)
	ss.check()

	if !ss.IsActive() {
		t.Error("screen saver should be active")
	}

	// Reset activity
	ss.ResetActivity()

	if ss.IsActive() {
		t.Error("screen saver should be deactivated after activity")
	}
}

func TestScreenSaverModes(t *testing.T) {
	modes := []struct {
		name                     string
		mode                     Mode
		expectedActiveBrightness uint8
	}{
		{"dim", ModeDim, 50},
		{"blank", ModeBlank, 0},
	}

	for _, tc := range modes {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{
				Enabled:          true,
				Mode:             tc.mode,
				IdleTimeout:      50 * time.Millisecond,
				DimBrightness:    50,
				NormalBrightness: 255,
			}

			disp := display.NewMockDisplay(128, 64)
			log := logger.NewDefault()

			ss := New(cfg, disp, log)

			// Trigger activation
			time.Sleep(100 * time.Millisecond)
			ss.check()

			if !ss.IsActive() {
				t.Error("screen saver should be active")
			}
		})
	}
}

func TestUpdateConfig(t *testing.T) {
	initialCfg := Config{
		Enabled:          true,
		Mode:             ModeDim,
		IdleTimeout:      30 * time.Second,
		DimBrightness:    50,
		NormalBrightness: 255,
	}

	disp := display.NewMockDisplay(128, 64)
	log := logger.NewDefault()

	ss := New(initialCfg, disp, log)

	// Update config
	newCfg := Config{
		Enabled:          true,
		Mode:             ModeBlank,
		IdleTimeout:      60 * time.Second,
		DimBrightness:    30,
		NormalBrightness: 200,
	}

	ss.UpdateConfig(newCfg)

	// Verify config was updated (check via reflection or behavior)
	// For now, just verify no crash
}

func TestStartStop(t *testing.T) {
	cfg := Config{
		Enabled:          true,
		Mode:             ModeDim,
		IdleTimeout:      1 * time.Second,
		DimBrightness:    50,
		NormalBrightness: 255,
	}

	disp := display.NewMockDisplay(128, 64)
	log := logger.NewDefault()

	ss := New(cfg, disp, log)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := ss.Start(ctx); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	ss.Stop()

	// Should not panic or error
}
