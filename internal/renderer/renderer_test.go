package renderer

import (
	"testing"

	"github.com/ausil/i2c-display/internal/config"
	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/stats"
)

func TestSystemPage(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)
	if err := disp.Init(); err != nil {
		t.Fatalf("failed to init display: %v", err)
	}

	page := NewSystemPage()

	testStats := &stats.SystemStats{
		Hostname:    "testhost",
		CPUTemp:     45.5,
		MemoryUsed:  2 * 1024 * 1024 * 1024,
		MemoryTotal: 4 * 1024 * 1024 * 1024,
		DiskUsed:    50 * 1024 * 1024 * 1024,
		DiskTotal:   100 * 1024 * 1024 * 1024,
	}

	if err := page.Render(disp, testStats); err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	// Verify some draw calls were made
	calls := disp.GetCalls()
	if len(calls) < 3 {
		t.Errorf("expected multiple draw calls, got %d", len(calls))
	}

	// Verify Show was called
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

func TestNetworkPage(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)
	if err := disp.Init(); err != nil {
		t.Fatalf("failed to init display: %v", err)
	}

	page := NewNetworkPage(1, 3, 5)

	testStats := &stats.SystemStats{
		Hostname: "testhost",
		Interfaces: []stats.NetInterface{
			{Name: "eth0", IPv4Addrs: []string{"192.168.1.100"}},
			{Name: "wlan0", IPv4Addrs: []string{"10.0.0.50"}},
			{Name: "usb0", IPv4Addrs: []string{"172.16.0.1"}},
		},
	}

	if err := page.Render(disp, testStats); err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	calls := disp.GetCalls()
	if len(calls) < 3 {
		t.Errorf("expected multiple draw calls, got %d", len(calls))
	}
}

func TestNetworkPageMultiplePages(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)

	// Create 6 interfaces, which should create 2 pages with max 3 per page
	interfaces := make([]stats.NetInterface, 6)
	for i := 0; i < 6; i++ {
		interfaces[i] = stats.NetInterface{
			Name:      "eth" + string(rune('0'+i)),
			IPv4Addrs: []string{"192.168.1." + string(rune('0'+i))},
		}
	}

	testStats := &stats.SystemStats{
		Hostname:   "testhost",
		Interfaces: interfaces,
	}

	// Test page 1
	page1 := NewNetworkPage(1, 3, 6)
	if page1.totalPages != 2 {
		t.Errorf("expected 2 total pages, got %d", page1.totalPages)
	}
	if page1.interfaceStartIdx != 0 || page1.interfaceEndIdx != 3 {
		t.Errorf("page 1 should show interfaces 0-3, got %d-%d", page1.interfaceStartIdx, page1.interfaceEndIdx)
	}

	if err := page1.Render(disp, testStats); err != nil {
		t.Fatalf("Render page 1 failed: %v", err)
	}

	// Test page 2
	page2 := NewNetworkPage(2, 3, 6)
	if page2.interfaceStartIdx != 3 || page2.interfaceEndIdx != 6 {
		t.Errorf("page 2 should show interfaces 3-6, got %d-%d", page2.interfaceStartIdx, page2.interfaceEndIdx)
	}

	if err := page2.Render(disp, testStats); err != nil {
		t.Fatalf("Render page 2 failed: %v", err)
	}
}

func TestRenderer(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)
	cfg := config.Default()

	renderer := NewRenderer(disp, cfg)

	testStats := &stats.SystemStats{
		Hostname:    "testhost",
		CPUTemp:     45.5,
		MemoryUsed:  2 * 1024 * 1024 * 1024,
		MemoryTotal: 4 * 1024 * 1024 * 1024,
		DiskUsed:    50 * 1024 * 1024 * 1024,
		DiskTotal:   100 * 1024 * 1024 * 1024,
		Interfaces: []stats.NetInterface{
			{Name: "eth0", IPv4Addrs: []string{"192.168.1.100"}},
			{Name: "wlan0", IPv4Addrs: []string{"10.0.0.50"}},
		},
	}

	// Build pages
	renderer.BuildPages(testStats)

	// Should have 1 system page + 1 network page (2 interfaces with max 3 per page)
	if renderer.PageCount() != 2 {
		t.Errorf("expected 2 pages, got %d", renderer.PageCount())
	}

	// Render first page
	if err := renderer.RenderPage(0, testStats); err != nil {
		t.Fatalf("RenderPage(0) failed: %v", err)
	}

	// Render second page
	if err := renderer.RenderPage(1, testStats); err != nil {
		t.Fatalf("RenderPage(1) failed: %v", err)
	}

	// Invalid page index should return an error
	if err := renderer.RenderPage(99, testStats); err == nil {
		t.Errorf("RenderPage(99) should return an error for out-of-range index")
	}
}

func TestRendererNoInterfaces(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)
	cfg := config.Default()

	renderer := NewRenderer(disp, cfg)

	testStats := &stats.SystemStats{
		Hostname:   "testhost",
		Interfaces: []stats.NetInterface{},
	}

	renderer.BuildPages(testStats)

	// Should have only 1 system page
	if renderer.PageCount() != 1 {
		t.Errorf("expected 1 page, got %d", renderer.PageCount())
	}
}

func TestTextHelpers(t *testing.T) {
	// Test TruncateText
	short := "short"
	truncated := TruncateText(short, 1000)
	if truncated != short {
		t.Errorf("short text should not be truncated")
	}

	long := "this is a very long string that should be truncated"
	truncated = TruncateText(long, 50)
	if len(truncated) > len(long) {
		t.Error("truncated text should not be longer than original")
	}
	if MeasureText(truncated) > 50 {
		t.Error("truncated text exceeds max width")
	}
}

func TestNewLayout(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		height      int
		wantHeader  bool
		wantSep     bool
		minLines    int
	}{
		{
			name:        "small 128x32",
			width:       128,
			height:      32,
			wantHeader:  true,
			wantSep:     true,
			minLines:    1,
		},
		{
			name:        "medium 128x64",
			width:       128,
			height:      64,
			wantHeader:  true,
			wantSep:     true,
			minLines:    3,
		},
		{
			name:        "large 256x128",
			width:       256,
			height:      128,
			wantHeader:  true,
			wantSep:     true,
			minLines:    6,
		},
		{
			name:        "tiny 96x16",
			width:       96,
			height:      16,
			wantHeader:  true,
			wantSep:     true,
			minLines:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bounds := display.NewMockDisplay(tt.width, tt.height).GetBounds()
			layout := NewLayout(bounds)

			if layout.ShowHeader != tt.wantHeader {
				t.Errorf("ShowHeader = %v, want %v", layout.ShowHeader, tt.wantHeader)
			}

			if layout.ShowSeparator != tt.wantSep {
				t.Errorf("ShowSeparator = %v, want %v", layout.ShowSeparator, tt.wantSep)
			}

			if len(layout.ContentLines) < tt.minLines {
				t.Errorf("ContentLines has %d lines, want at least %d",
					len(layout.ContentLines), tt.minLines)
			}

			if layout.MaxContentLines < tt.minLines {
				t.Errorf("MaxContentLines = %d, want at least %d",
					layout.MaxContentLines, tt.minLines)
			}
		})
	}
}

func TestCenterText(t *testing.T) {
	tests := []struct {
		text  string
		width int
	}{
		{"Hello", 128},
		{"X", 128},
		{"Long text that is quite wide", 256}, // Use wider display for long text
		{"Test", 64},
		{"", 128},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			x := CenterText(tt.text, tt.width)

			textWidth := MeasureText(tt.text)

			// If text fits within display, x should be non-negative
			if textWidth < tt.width {
				if x < 0 {
					t.Errorf("CenterText returned negative x: %d for text that fits", x)
				}
			}
			// Negative x is okay if text is too wide to center
		})
	}
}

func TestPageTitles(t *testing.T) {
	t.Run("network page title", func(t *testing.T) {
		page := NewNetworkPage(1, 3, 9)
		title := page.Title()
		if title == "" {
			t.Error("expected non-empty title")
		}
		// Should contain page number
		if title != "Network 1/3" {
			t.Errorf("expected 'Network 1/3', got %q", title)
		}
	})

	t.Run("system page title", func(t *testing.T) {
		page := NewSystemPage()
		title := page.Title()
		if title == "" {
			t.Error("expected non-empty title")
		}
		if title != "System" {
			t.Errorf("expected 'System', got %q", title)
		}
	})

	t.Run("system metric page titles", func(t *testing.T) {
		tests := []struct {
			metric SystemMetricType
			want   string
		}{
			{SystemMetricDisk, "Disk"},
			{SystemMetricMemory, "Memory"},
			{SystemMetricCPU, "CPU"},
			{SystemMetricAll, "System"},
		}

		for _, tt := range tests {
			page := NewSystemPageForMetric(tt.metric)
			title := page.Title()
			if title != tt.want {
				t.Errorf("metric %d: expected title %q, got %q", tt.metric, tt.want, title)
			}
		}
	})
}

func TestGetPages(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)
	cfg := config.Default()
	renderer := NewRenderer(disp, cfg)

	testStats := &stats.SystemStats{
		Hostname: "testhost",
		Interfaces: []stats.NetInterface{
			{Name: "eth0", IPv4Addrs: []string{"192.168.1.100"}},
		},
	}

	renderer.BuildPages(testStats)
	pages := renderer.GetPages()

	if len(pages) != renderer.PageCount() {
		t.Errorf("GetPages returned %d pages, PageCount is %d",
			len(pages), renderer.PageCount())
	}

	if len(pages) == 0 {
		t.Error("expected at least one page")
	}
}

func TestSystemPageMetricTypes(t *testing.T) {
	disp := display.NewMockDisplay(128, 32)

	testStats := &stats.SystemStats{
		Hostname:    "testhost",
		CPUTemp:     55.5,
		MemoryUsed:  2 * 1024 * 1024 * 1024,
		MemoryTotal: 4 * 1024 * 1024 * 1024,
		DiskUsed:    50 * 1024 * 1024 * 1024,
		DiskTotal:   100 * 1024 * 1024 * 1024,
	}

	metrics := []SystemMetricType{
		SystemMetricDisk,
		SystemMetricMemory,
		SystemMetricCPU,
		SystemMetricAll,
	}

	for _, metric := range metrics {
		t.Run(string(rune(metric)), func(t *testing.T) {
			page := NewSystemPageForMetric(metric)
			if err := page.Render(disp, testStats); err != nil {
				t.Errorf("Render failed for metric %d: %v", metric, err)
			}
		})
	}
}

func TestSystemPageZeroCPUTemp(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)

	testStats := &stats.SystemStats{
		Hostname:    "testhost",
		CPUTemp:     0, // Zero temp should skip CPU line
		MemoryUsed:  2 * 1024 * 1024 * 1024,
		MemoryTotal: 4 * 1024 * 1024 * 1024,
		DiskUsed:    50 * 1024 * 1024 * 1024,
		DiskTotal:   100 * 1024 * 1024 * 1024,
	}

	page := NewSystemPage()
	if err := page.Render(disp, testStats); err != nil {
		t.Errorf("Render with zero CPU temp failed: %v", err)
	}
}

func TestNetworkPageIPv6(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)

	testStats := &stats.SystemStats{
		Hostname: "testhost",
		Interfaces: []stats.NetInterface{
			{
				Name:       "eth0",
				IPv4Addrs:  []string{},
				IPv6Addrs:  []string{"fe80::1234:5678:90ab:cdef"},
			},
			{
				Name:       "eth1",
				IPv4Addrs:  []string{"192.168.1.100"},
				IPv6Addrs:  []string{"fe80::abcd:ef12:3456:7890"},
			},
		},
	}

	page := NewNetworkPage(1, 3, 2)
	if err := page.Render(disp, testStats); err != nil {
		t.Errorf("Render with IPv6 failed: %v", err)
	}
}

func TestNetworkPageNoAddress(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)

	testStats := &stats.SystemStats{
		Hostname: "testhost",
		Interfaces: []stats.NetInterface{
			{
				Name:       "eth0",
				IPv4Addrs:  []string{},
				IPv6Addrs:  []string{},
			},
		},
	}

	page := NewNetworkPage(1, 3, 1)
	if err := page.Render(disp, testStats); err != nil {
		t.Errorf("Render with no address failed: %v", err)
	}
}

func TestRendererSmallDisplay(t *testing.T) {
	disp := display.NewMockDisplay(128, 32)
	cfg := config.Default()
	renderer := NewRenderer(disp, cfg)

	testStats := &stats.SystemStats{
		Hostname:    "testhost",
		CPUTemp:     55.5,
		MemoryUsed:  2 * 1024 * 1024 * 1024,
		MemoryTotal: 4 * 1024 * 1024 * 1024,
		DiskUsed:    50 * 1024 * 1024 * 1024,
		DiskTotal:   100 * 1024 * 1024 * 1024,
		Interfaces: []stats.NetInterface{
			{Name: "eth0", IPv4Addrs: []string{"192.168.1.100"}},
		},
	}

	renderer.BuildPages(testStats)

	// Small displays should create separate metric pages
	if renderer.PageCount() < 3 {
		t.Errorf("expected at least 3 pages for small display, got %d", renderer.PageCount())
	}

	// Render all pages
	for i := 0; i < renderer.PageCount(); i++ {
		if err := renderer.RenderPage(i, testStats); err != nil {
			t.Errorf("RenderPage(%d) failed: %v", i, err)
		}
	}
}

func TestDrawHelpers(t *testing.T) {
	disp := display.NewMockDisplay(128, 64)

	t.Run("DrawText", func(t *testing.T) {
		if err := DrawText(disp, 0, 0, "Hello"); err != nil {
			t.Errorf("DrawText failed: %v", err)
		}
	})

	t.Run("DrawTextCentered", func(t *testing.T) {
		if err := DrawTextCentered(disp, 32, "Centered"); err != nil {
			t.Errorf("DrawTextCentered failed: %v", err)
		}
	})

	t.Run("DrawLine", func(t *testing.T) {
		if err := DrawLine(disp, 10); err != nil {
			t.Errorf("DrawLine failed: %v", err)
		}
	})
}

func TestMeasureTextEmpty(t *testing.T) {
	width := MeasureText("")
	if width != 0 {
		t.Errorf("expected width 0 for empty string, got %d", width)
	}
}
