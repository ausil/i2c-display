package renderer

import (
	"testing"

	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/stats"
)

func TestLoadGraphPageTitle(t *testing.T) {
	page := NewLoadGraphPage(0)
	if title := page.Title(); title != "Load" {
		t.Errorf("expected title 'Load', got %q", title)
	}
}

func TestLoadGraphPageRender(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)
	page := NewLoadGraphPage(0)

	testStats := &stats.SystemStats{
		Hostname:  "testhost",
		LoadAvg1:  0.50,
		LoadAvg5:  0.40,
		LoadAvg15: 0.35,
		NumCPU:    4,
	}

	if err := page.Render(disp, testStats); err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	calls := disp.GetCalls()
	foundShow := false
	for _, call := range calls {
		if call == "Show" {
			foundShow = true
			break
		}
	}
	if !foundShow {
		t.Error("expected Show() to be called")
	}
}

func TestLoadGraphPageRenderSmall(t *testing.T) {
	disp := display.NewMockDisplay(128, 32)
	page := NewLoadGraphPage(0)

	testStats := &stats.SystemStats{
		Hostname:  "testhost",
		LoadAvg1:  0.25,
		LoadAvg5:  0.21,
		LoadAvg15: 0.27,
		NumCPU:    4,
	}

	if err := page.Render(disp, testStats); err != nil {
		t.Fatalf("Render() on small display failed: %v", err)
	}

	calls := disp.GetCalls()
	foundShow := false
	for _, call := range calls {
		if call == "Show" {
			foundShow = true
			break
		}
	}
	if !foundShow {
		t.Error("expected Show() to be called")
	}
}

func TestLoadGraphPageHistory(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)
	page := NewLoadGraphPage(0)

	// Render multiple times to accumulate history
	for i := 0; i < 10; i++ {
		testStats := &stats.SystemStats{
			Hostname:  "testhost",
			LoadAvg1:  float64(i) * 0.1,
			LoadAvg5:  0.20,
			LoadAvg15: 0.30,
			NumCPU:    4,
		}
		if err := page.Render(disp, testStats); err != nil {
			t.Fatalf("Render() iteration %d failed: %v", i, err)
		}
	}

	if page.count != 10 {
		t.Errorf("expected count=10, got %d", page.count)
	}

	// Verify ring buffer wrapping: fill beyond capacity
	for i := 0; i < loadHistorySize+5; i++ {
		testStats := &stats.SystemStats{
			Hostname:  "testhost",
			LoadAvg1:  1.0,
			LoadAvg5:  0.50,
			LoadAvg15: 0.40,
			NumCPU:    4,
		}
		if err := page.Render(disp, testStats); err != nil {
			t.Fatalf("Render() wrap iteration %d failed: %v", i, err)
		}
	}

	// Count should be capped at loadHistorySize
	if page.count != loadHistorySize {
		t.Errorf("expected count=%d after wrapping, got %d", loadHistorySize, page.count)
	}
}

func TestLoadGraphPageZeroLoad(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)
	page := NewLoadGraphPage(0)

	testStats := &stats.SystemStats{
		Hostname:  "testhost",
		LoadAvg1:  0,
		LoadAvg5:  0,
		LoadAvg15: 0,
		NumCPU:    4,
	}

	if err := page.Render(disp, testStats); err != nil {
		t.Fatalf("Render() with zero load failed: %v", err)
	}
}

func TestLoadGraphPageGetSamples(t *testing.T) {
	page := NewLoadGraphPage(0)

	// Empty history
	samples := page.getSamples()
	if len(samples) != 0 {
		t.Errorf("expected 0 samples from empty history, got %d", len(samples))
	}

	// Add a few samples
	page.history[0] = 1.0
	page.history[1] = 2.0
	page.history[2] = 3.0
	page.head = 3
	page.count = 3

	samples = page.getSamples()
	if len(samples) != 3 {
		t.Fatalf("expected 3 samples, got %d", len(samples))
	}
	if samples[0] != 1.0 || samples[1] != 2.0 || samples[2] != 3.0 {
		t.Errorf("unexpected samples: %v", samples)
	}
}
