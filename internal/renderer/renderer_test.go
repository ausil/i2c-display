package renderer

import (
	"testing"

	"github.com/denniskorablev/ssd1306-display/internal/config"
	"github.com/denniskorablev/ssd1306-display/internal/display"
	"github.com/denniskorablev/ssd1306-display/internal/stats"
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

	// Invalid page index should not error
	if err := renderer.RenderPage(99, testStats); err != nil {
		t.Errorf("RenderPage(99) should not error, got %v", err)
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
