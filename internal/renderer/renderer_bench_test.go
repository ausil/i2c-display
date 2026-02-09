package renderer

import (
	"testing"

	"github.com/ausil/i2c-display/internal/config"
	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/stats"
)

func BenchmarkRenderPage(b *testing.B) {
	cfg := config.Default()
	disp := display.NewMockDisplay(128, 64)
	rend := NewRenderer(disp, cfg)

	testStats := &stats.SystemStats{
		Hostname:    "testhost",
		CPUTemp:     45.2,
		MemoryUsed:  2684354560,  // ~2.5GB in bytes
		MemoryTotal: 4294967296,  // 4GB in bytes
		DiskUsed:    13421772800, // ~12.5GB in bytes
		DiskTotal:   29629292544, // ~27.6GB in bytes
		Interfaces:  []stats.NetInterface{},
	}

	rend.BuildPages(testStats)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := rend.pages[0].Render(disp, testStats); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildPages(b *testing.B) {
	cfg := config.Default()
	disp := display.NewMockDisplay(128, 64)
	rend := NewRenderer(disp, cfg)

	testStats := &stats.SystemStats{
		Hostname:    "testhost",
		CPUTemp:     45.2,
		MemoryUsed:  2684354560,
		MemoryTotal: 4294967296,
		DiskUsed:    13421772800,
		DiskTotal:   29629292544,
		Interfaces: []stats.NetInterface{
			{Name: "eth0", IPv4Addrs: []string{"192.168.1.100"}},
			{Name: "wlan0", IPv4Addrs: []string{"10.0.0.50"}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rend.BuildPages(testStats)
	}
}
