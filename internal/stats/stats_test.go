package stats

import (
	"testing"

	"github.com/ausil/i2c-display/internal/config"
)

func TestSystemStats(t *testing.T) {
	stats := &SystemStats{
		MemoryUsed:  2 * 1024 * 1024 * 1024, // 2 GB
		MemoryTotal: 4 * 1024 * 1024 * 1024, // 4 GB
		DiskUsed:    50 * 1024 * 1024 * 1024, // 50 GB
		DiskTotal:   100 * 1024 * 1024 * 1024, // 100 GB
	}

	// Test percentages
	if stats.MemoryPercent() != 50.0 {
		t.Errorf("expected MemoryPercent=50.0, got %f", stats.MemoryPercent())
	}

	if stats.DiskPercent() != 50.0 {
		t.Errorf("expected DiskPercent=50.0, got %f", stats.DiskPercent())
	}

	// Test GB conversions
	if stats.MemoryUsedGB() < 1.99 || stats.MemoryUsedGB() > 2.01 {
		t.Errorf("expected MemoryUsedGB~2.0, got %f", stats.MemoryUsedGB())
	}

	if stats.MemoryTotalGB() < 3.99 || stats.MemoryTotalGB() > 4.01 {
		t.Errorf("expected MemoryTotalGB~4.0, got %f", stats.MemoryTotalGB())
	}

	if stats.DiskUsedGB() < 49.9 || stats.DiskUsedGB() > 50.1 {
		t.Errorf("expected DiskUsedGB~50.0, got %f", stats.DiskUsedGB())
	}

	if stats.DiskTotalGB() < 99.9 || stats.DiskTotalGB() > 100.1 {
		t.Errorf("expected DiskTotalGB~100.0, got %f", stats.DiskTotalGB())
	}
}

func TestSystemStatsZeroTotal(t *testing.T) {
	stats := &SystemStats{
		MemoryUsed:  0,
		MemoryTotal: 0,
		DiskUsed:    0,
		DiskTotal:   0,
	}

	if stats.MemoryPercent() != 0 {
		t.Errorf("expected MemoryPercent=0 for zero total, got %f", stats.MemoryPercent())
	}

	if stats.DiskPercent() != 0 {
		t.Errorf("expected DiskPercent=0 for zero total, got %f", stats.DiskPercent())
	}
}

func TestCPUTempCollector(t *testing.T) {
	collector := NewCPUTempCollector("../../testdata/sys/class/thermal/thermal_zone0/temp")

	temp, err := collector.GetTemperature()
	if err != nil {
		t.Fatalf("GetTemperature() failed: %v", err)
	}

	// Should be 45.2°C (45200 millidegrees)
	expected := 45.2
	if temp < expected-0.1 || temp > expected+0.1 {
		t.Errorf("expected temp~%.1f, got %.1f", expected, temp)
	}
}

func TestCPUTempCollectorNonExistent(t *testing.T) {
	collector := NewCPUTempCollector("/nonexistent/path")

	_, err := collector.GetTemperature()
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestMemoryCollector(t *testing.T) {
	collector := NewMemoryCollectorWithPath("../../testdata/proc/meminfo")

	used, total, err := collector.GetMemory()
	if err != nil {
		t.Fatalf("GetMemory() failed: %v", err)
	}

	// Total should be 4048384 kB = ~4 GB
	expectedTotal := uint64(4048384 * 1024)
	if total != expectedTotal {
		t.Errorf("expected total=%d, got %d", expectedTotal, total)
	}

	// Used = Total - Available = 4048384 - 1536000 = 2512384 kB
	expectedUsed := uint64((4048384 - 1536000) * 1024)
	if used != expectedUsed {
		t.Errorf("expected used=%d, got %d", expectedUsed, used)
	}
}

func TestMemoryCollectorNonExistent(t *testing.T) {
	collector := NewMemoryCollectorWithPath("/nonexistent/meminfo")

	_, _, err := collector.GetMemory()
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestDiskCollector(t *testing.T) {
	collector := NewDiskCollector("/")

	used, total, err := collector.GetDisk()
	if err != nil {
		t.Fatalf("GetDisk() failed: %v", err)
	}

	if total == 0 {
		t.Error("expected non-zero total disk space")
	}

	if used == 0 {
		t.Error("expected non-zero used disk space")
	}

	if used > total {
		t.Errorf("used (%d) should not exceed total (%d)", used, total)
	}
}

func TestDiskCollectorNonExistent(t *testing.T) {
	collector := NewDiskCollector("/nonexistent/path/that/does/not/exist")

	_, _, err := collector.GetDisk()
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestNetworkCollector(t *testing.T) {
	cfg := config.NetworkConfig{
		AutoDetect: true,
		InterfaceFilter: config.InterfaceFilter{
			Include: []string{"*"},
			Exclude: []string{"lo"},
		},
		ShowIPv4: true,
		ShowIPv6: false,
	}

	collector := NewNetworkCollector(cfg)

	interfaces, err := collector.GetInterfaces()
	if err != nil {
		t.Fatalf("GetInterfaces() failed: %v", err)
	}

	// Should have at least some interfaces (results will vary by system)
	// Just verify the structure is correct
	for _, iface := range interfaces {
		if iface.Name == "" {
			t.Error("interface should have a name")
		}

		// loopback should be excluded
		if iface.Name == "lo" {
			t.Error("loopback interface should be excluded")
		}

		// Should have IPv4 addresses since ShowIPv4 is true
		if cfg.ShowIPv4 && len(iface.IPv4Addrs) == 0 {
			t.Logf("warning: interface %s has no IPv4 addresses", iface.Name)
		}
	}
}

func TestNetworkCollectorFiltering(t *testing.T) {
	tests := []struct {
		name        string
		config      config.NetworkConfig
		ifaceName   string
		shouldInclude bool
	}{
		{
			name: "exclude loopback",
			config: config.NetworkConfig{
				AutoDetect: true,
				InterfaceFilter: config.InterfaceFilter{
					Exclude: []string{"lo"},
				},
			},
			ifaceName:   "lo",
			shouldInclude: false,
		},
		{
			name: "exclude docker wildcard",
			config: config.NetworkConfig{
				AutoDetect: true,
				InterfaceFilter: config.InterfaceFilter{
					Exclude: []string{"docker*"},
				},
			},
			ifaceName:   "docker0",
			shouldInclude: false,
		},
		{
			name: "include eth0",
			config: config.NetworkConfig{
				AutoDetect: false,
				InterfaceFilter: config.InterfaceFilter{
					Include: []string{"eth0"},
				},
			},
			ifaceName:   "eth0",
			shouldInclude: true,
		},
		{
			name: "include with wildcard",
			config: config.NetworkConfig{
				AutoDetect: false,
				InterfaceFilter: config.InterfaceFilter{
					Include: []string{"eth*"},
				},
			},
			ifaceName:   "eth0",
			shouldInclude: true,
		},
		{
			name: "auto detect with no filters",
			config: config.NetworkConfig{
				AutoDetect: true,
				InterfaceFilter: config.InterfaceFilter{},
			},
			ifaceName:   "eth0",
			shouldInclude: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewNetworkCollector(tt.config)
			result := collector.shouldInclude(tt.ifaceName)
			if result != tt.shouldInclude {
				t.Errorf("shouldInclude(%s) = %v, want %v", tt.ifaceName, result, tt.shouldInclude)
			}
		})
	}
}
