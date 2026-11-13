package renderer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/stats"
)

// panelWidth is the usable width on a 128 px panel, i.e. the bounds minus the
// left and right margins that every page draws inside.
const panelWidth = 128 - 2*MarginLeft

func TestFitScaleKeepsRealValuesUntruncated(t *testing.T) {
	// Values seen on real hardware; each one used to be cut off on a 128x64
	// panel because the full-size 7x13 face only fits 18 characters.
	tests := []struct {
		name string
		line string
	}{
		{"IPv4 with space after colon", networkLine("end0", "172.16.46.134", false)},
		{"IPv4 on a longer interface name", networkLine("wlan0", "172.16.46.134", false)},
		{"load averages", loadLabel(&stats.SystemStats{LoadAvg1: 0.52, LoadAvg5: 0.48, LoadAvg15: 0.40})},
		{"disk usage", fmt.Sprintf("%.0f%% %.1f/%.1fG", 45.2, 12.3, 465.7)},
		{"memory usage", fmt.Sprintf("%.0f%% %.1f/%.1fG", 62.4, 5.1, 15.6)},
		{"disk usage, large values", fmt.Sprintf("%.0f%% %.1f/%.1fG", 100.0, 931.5, 931.5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scale := fitScale([]string{tt.line}, panelWidth, 0)
			if got := truncateAt(tt.line, panelWidth, scale); got != tt.line {
				t.Errorf("truncated %q to %q (full-size %d px, compact %d px, budget %d px)",
					tt.line, got, MeasureText(tt.line), MeasureTextSmall(tt.line), panelWidth)
			}
		})
	}
}

func TestLoadLabelShowsAllThreeAverages(t *testing.T) {
	s := &stats.SystemStats{LoadAvg1: 0.52, LoadAvg5: 0.48, LoadAvg15: 0.4}

	label := loadLabel(s)
	for _, want := range []string{"1m", "5m", "15m", "0.40"} {
		if !strings.Contains(label, want) {
			t.Errorf("load label %q is missing %q", label, want)
		}
	}
	if MeasureTextSmall(label) > panelWidth {
		t.Errorf("load label %q needs %d px in the compact face, budget is %d px",
			label, MeasureTextSmall(label), panelWidth)
	}
}

func TestNetworkLineFormats(t *testing.T) {
	if got, want := networkLine("end0", "172.16.46.134", false), "end0: 172.16.46.134"; got != want {
		t.Errorf("standard format = %q, want %q", got, want)
	}
	if got, want := networkLine("end0", "172.16.46.134", true), "end0:172.16.46.134"; got != want {
		t.Errorf("compact format = %q, want %q", got, want)
	}
}

func TestFitScalePrefersFullSizeFont(t *testing.T) {
	lines := []string{"malek", "D:45% 12.3/465.7G", "Page 1/2"}
	if got := fitScale(lines, panelWidth, 0); got != 1 {
		t.Errorf("fitScale() = %v, want 1 for lines that fit at full size", got)
	}
}

func TestFitScaleDropsToCompactFace(t *testing.T) {
	if got := fitScale([]string{networkLine("end0", "172.16.46.134", false)}, panelWidth, 0); got != 0.5 {
		t.Errorf("fitScale() = %v, want 0.5 for a line that only fits compact", got)
	}
}

func TestFitScaleNeverUpscalesAHalfScaleLayout(t *testing.T) {
	if got := fitScale([]string{"malek"}, panelWidth, 0.5); got != 0.5 {
		t.Errorf("fitScale() = %v, want 0.5 to preserve the layout's own scale", got)
	}
}

// compareBand reports whether two displays show identical pixels in the given
// row band and column range.
func compareBand(t *testing.T, got, want *display.MockDisplay, y0, y1, x0, x1 int) bool {
	t.Helper()
	same := true
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if got.GetPixel(x, y) != want.GetPixel(x, y) {
				if same {
					t.Errorf("first difference at (%d,%d): panel=%v reference=%v", x, y, got.GetPixel(x, y), want.GetPixel(x, y))
				}
				same = false
			}
		}
	}
	return same
}

// Standard 128x64 content rows, per NewLayout.
var standardRows = []int{16, 28, 40}

func TestNetworkPageShowsFullIPv4OnPanel(t *testing.T) {
	s := &stats.SystemStats{
		Hostname:   "malek",
		Interfaces: []stats.NetInterface{{Name: "end0", IPv4Addrs: []string{"172.16.46.134"}}},
	}

	page := NewNetworkPage(1, 3, 1, 0)
	got := display.NewMockDisplay(128, 64)
	if err := page.Render(got, s); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Reference: the full address drawn with the compact face at the same spot.
	// Before the fit fix this row held the full-size font truncated to
	// "end0: 172.16.46...", which lost the last octet.
	want := display.NewMockDisplay(128, 64)
	if err := DrawTextColorScaled(want, MarginLeft, standardRows[0], "end0: 172.16.46.134", ColorGreen, 0.5); err != nil {
		t.Fatalf("reference draw failed: %v", err)
	}

	if !compareBand(t, got, want, standardRows[0], standardRows[0]+ScaledTextHeight(0.5), 0, 128) {
		t.Errorf("network page does not show the full IPv4 address")
	}
}

func TestLoadGraphPageShowsAllThreeAveragesOnPanel(t *testing.T) {
	s := &stats.SystemStats{Hostname: "malek", LoadAvg1: 0.52, LoadAvg5: 0.48, LoadAvg15: 0.4, NumCPU: 8}

	page := NewLoadGraphPage(1)
	got := display.NewMockDisplay(128, 64)
	if err := page.Render(got, s); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	want := display.NewMockDisplay(128, 64)
	if err := DrawTextColorScaled(want, MarginLeft, standardRows[0], loadLabel(s), LoadColor(s.LoadAvg1, s.NumCPU), 0.5); err != nil {
		t.Fatalf("reference draw failed: %v", err)
	}

	if !compareBand(t, got, want, standardRows[0], standardRows[0]+ScaledTextHeight(0.5), 0, 128) {
		t.Errorf("load page does not show all three load averages")
	}
}

func TestSystemPageShowsDiskAndMemoryTotalsOnPanel(t *testing.T) {
	s := &stats.SystemStats{
		Hostname: "malek",
		DiskUsed: 210e9, DiskTotal: 4657e8,
		MemoryUsed: 32e8, MemoryTotal: 156e8,
	}

	page := NewSystemPage(0)
	got := display.NewMockDisplay(128, 64)
	if err := page.Render(got, s); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	textX := MarginLeft + IconWidth + IconGap
	iconBudget := 128 - 2*MarginLeft - IconWidth - IconGap
	for i, wantText := range []string{
		fmt.Sprintf("%.0f%% %.1f/%.1fG", s.DiskPercent(), s.DiskUsedGB(), s.DiskTotalGB()),
		fmt.Sprintf("%.0f%% %.1f/%.1fG", s.MemoryPercent(), s.MemoryUsedGB(), s.MemoryTotalGB()),
	} {
		if !strings.Contains(wantText, "/") {
			t.Fatalf("fixture %q has no used/total pair to show", wantText)
		}
		if MeasureText(wantText) > iconBudget {
			t.Fatalf("fixture %q needs %d px at full size but the budget is %d px",
				wantText, MeasureText(wantText), iconBudget)
		}

		want := display.NewMockDisplay(128, 64)
		if err := DrawTextColorScaled(want, textX, standardRows[i], wantText, MetricColor(45), 1); err != nil {
			t.Fatalf("reference draw failed: %v", err)
		}
		if !compareBand(t, got, want, standardRows[i], standardRows[i]+ScaledTextHeight(1), textX, 128) {
			t.Errorf("system page does not show the disk/memory total %q", wantText)
		}
	}
}
