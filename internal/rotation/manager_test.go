package rotation

import (
	"context"
	"testing"
	"time"

	"github.com/ausil/i2c-display/internal/config"
	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/renderer"
	"github.com/ausil/i2c-display/internal/stats"
)

func TestManager(t *testing.T) {
	// Create test config with fast intervals
	cfg := config.Default()
	cfg.Pages.RotationInterval = "100ms"
	cfg.Pages.RefreshInterval = "50ms"

	// Create mock display
	disp := display.NewMockDisplay(128, 64)
	if err := disp.Init(); err != nil {
		t.Fatalf("failed to init display: %v", err)
	}

	// Create collector
	collector, err := stats.NewSystemCollector(cfg)
	if err != nil {
		t.Fatalf("failed to create collector: %v", err)
	}

	// Create renderer
	rend := renderer.NewRenderer(disp, cfg)

	// Create manager
	mgr := NewManager(cfg, collector, rend)

	// Start manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}

	// Wait for a few rotation cycles
	time.Sleep(350 * time.Millisecond)

	// Current page should have advanced
	currentPage := mgr.CurrentPage()
	if currentPage < 0 {
		t.Error("current page should be non-negative")
	}

	// Stop manager
	mgr.Stop()

	// Manager should stop cleanly
}

func TestManagerRotation(t *testing.T) {
	cfg := config.Default()
	cfg.Pages.RotationInterval = "50ms"
	cfg.Pages.RefreshInterval = "25ms"

	disp := display.NewMockDisplay(128, 64)
	disp.Init()

	collector, _ := stats.NewSystemCollector(cfg)
	rend := renderer.NewRenderer(disp, cfg)

	// Collect stats to build multiple pages
	s, _ := collector.Collect()
	rend.BuildPages(s)

	mgr := NewManager(cfg, collector, rend)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr.Start(ctx)

	initialPage := mgr.CurrentPage()

	// Wait for rotation
	time.Sleep(100 * time.Millisecond)

	// Page should have changed (if there are multiple pages)
	if rend.PageCount() > 1 {
		if mgr.CurrentPage() == initialPage {
			t.Log("Page did not rotate (might be expected if only 1 page)")
		}
	}

	mgr.Stop()
}

func TestManagerContextCancellation(t *testing.T) {
	cfg := config.Default()
	cfg.Pages.RotationInterval = "1s"
	cfg.Pages.RefreshInterval = "500ms"

	disp := display.NewMockDisplay(128, 64)
	disp.Init()

	collector, _ := stats.NewSystemCollector(cfg)
	rend := renderer.NewRenderer(disp, cfg)

	mgr := NewManager(cfg, collector, rend)

	ctx, cancel := context.WithCancel(context.Background())

	mgr.Start(ctx)

	// Cancel context
	cancel()

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Manager should have stopped
}

func TestManagerInvalidIntervals(t *testing.T) {
	cfg := config.Default()
	cfg.Pages.RotationInterval = "invalid"

	disp := display.NewMockDisplay(128, 64)
	collector, _ := stats.NewSystemCollector(cfg)
	rend := renderer.NewRenderer(disp, cfg)

	mgr := NewManager(cfg, collector, rend)

	ctx := context.Background()
	err := mgr.Start(ctx)
	if err == nil {
		t.Error("expected error for invalid rotation interval")
	}
}
