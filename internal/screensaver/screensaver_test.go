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

func TestWakeSuppressesScreensaver(t *testing.T) {
	cfg := Config{
		Enabled:          true,
		Mode:             ModeDim,
		IdleTimeout:      50 * time.Millisecond,
		DimBrightness:    50,
		NormalBrightness: 255,
		WakeDuration:     5 * time.Second,
	}

	disp := display.NewMockDisplay(128, 64)
	log := logger.NewDefault()
	ss := New(cfg, disp, log)

	// Activate the screensaver first
	time.Sleep(100 * time.Millisecond)
	ss.check()
	if !ss.IsActive() {
		t.Fatal("screensaver should be active before wake")
	}

	// Wake should deactivate immediately
	ss.Wake()
	if ss.IsActive() {
		t.Error("screensaver should be deactivated after Wake()")
	}

	// check() during the wake window should not re-activate
	ss.check()
	if ss.IsActive() {
		t.Error("screensaver should remain off during wake window")
	}
}

func TestWakeDisabledIsNoop(t *testing.T) {
	cfg := Config{Enabled: false}
	disp := display.NewMockDisplay(128, 64)
	log := logger.NewDefault()
	ss := New(cfg, disp, log)

	// Should not panic
	ss.Wake()
}

func TestInActiveHours(t *testing.T) {
	makeTime := func(h, m int) time.Time {
		return time.Date(2024, 1, 1, h, m, 0, 0, time.Local)
	}

	tests := []struct {
		name     string
		start    string
		end      string
		t        time.Time
		expected bool
	}{
		// Same-day range 08:00-22:00
		{"same-day: before start", "08:00", "22:00", makeTime(7, 59), false},
		{"same-day: at start", "08:00", "22:00", makeTime(8, 0), true},
		{"same-day: midday", "08:00", "22:00", makeTime(14, 0), true},
		{"same-day: at end", "08:00", "22:00", makeTime(22, 0), false},
		{"same-day: after end", "08:00", "22:00", makeTime(23, 0), false},
		// Overnight range 22:00-06:00
		{"overnight: before start", "22:00", "06:00", makeTime(21, 59), false},
		{"overnight: at start", "22:00", "06:00", makeTime(22, 0), true},
		{"overnight: midnight", "22:00", "06:00", makeTime(0, 0), true},
		{"overnight: before end", "22:00", "06:00", makeTime(5, 59), true},
		{"overnight: at end", "22:00", "06:00", makeTime(6, 0), false},
		{"overnight: midday", "22:00", "06:00", makeTime(12, 0), false},
		// Equal bounds = always active
		{"equal: always active", "12:00", "12:00", makeTime(8, 0), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ss := &ScreenSaver{
				cfg: Config{
					ActiveHours: ActiveHours{
						Enabled: true,
						Start:   tc.start,
						End:     tc.end,
					},
				},
			}
			got := ss.inActiveHours(tc.t)
			if got != tc.expected {
				t.Errorf("inActiveHours(%s) with start=%s end=%s: got %v, want %v",
					tc.t.Format("15:04"), tc.start, tc.end, got, tc.expected)
			}
		})
	}
}

func TestActiveHoursSuppressesScreensaver(t *testing.T) {
	cfg := Config{
		Enabled:          true,
		Mode:             ModeDim,
		IdleTimeout:      50 * time.Millisecond,
		DimBrightness:    50,
		NormalBrightness: 255,
		ActiveHours: ActiveHours{
			Enabled: true,
			Start:   "00:00",
			End:     "23:59",
		},
	}

	disp := display.NewMockDisplay(128, 64)
	log := logger.NewDefault()
	ss := New(cfg, disp, log)

	// Even after idle timeout would have fired, screensaver should stay off
	time.Sleep(100 * time.Millisecond)
	ss.check()

	if ss.IsActive() {
		t.Error("screensaver should be suppressed during active hours")
	}
}

func TestActiveHoursActivatesOutsideWindow(t *testing.T) {
	// Use a window that is definitely not now: 00:01-00:02
	// (almost certainly not the current time during a test run)
	cfg := Config{
		Enabled:          true,
		Mode:             ModeDim,
		DimBrightness:    50,
		NormalBrightness: 255,
		ActiveHours: ActiveHours{
			Enabled: true,
			Start:   "00:01",
			End:     "00:02",
		},
	}

	disp := display.NewMockDisplay(128, 64)
	log := logger.NewDefault()
	ss := New(cfg, disp, log)

	now := time.Now()
	// Skip if we happen to be running at 00:01
	if now.Hour() == 0 && now.Minute() == 1 {
		t.Skip("skipping: test running exactly during the 1-minute active window")
	}

	ss.check()

	if !ss.IsActive() {
		t.Error("screensaver should be active outside active hours window")
	}
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
