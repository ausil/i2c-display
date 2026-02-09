package stats

import (
	"fmt"
	"os"
	"strings"

	"github.com/denniskorablev/ssd1306-display/internal/config"
)

// SystemCollector collects all system statistics
type SystemCollector struct {
	config        *config.Config
	cpuCollector  *CPUTempCollector
	memCollector  *MemoryCollector
	diskCollector *DiskCollector
	netCollector  *NetworkCollector
	hostname      string
}

// NewSystemCollector creates a new system collector
func NewSystemCollector(cfg *config.Config) (*SystemCollector, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	// Extract short hostname if configured
	if cfg.SystemInfo.HostnameDisplay == "short" {
		if idx := strings.Index(hostname, "."); idx != -1 {
			hostname = hostname[:idx]
		}
	}

	return &SystemCollector{
		config:        cfg,
		cpuCollector:  NewCPUTempCollector(cfg.SystemInfo.TemperatureSource),
		memCollector:  NewMemoryCollector(),
		diskCollector: NewDiskCollector(cfg.SystemInfo.DiskPath),
		netCollector:  NewNetworkCollector(cfg.Network),
		hostname:      hostname,
	}, nil
}

// Collect gathers all system statistics
func (sc *SystemCollector) Collect() (*SystemStats, error) {
	stats := &SystemStats{
		Hostname: sc.hostname,
	}

	// Collect CPU temperature
	temp, err := sc.cpuCollector.GetTemperature()
	if err != nil {
		// Log warning but continue - temperature might not be available
		stats.CPUTemp = 0
	} else {
		// Convert to configured unit
		if sc.config.SystemInfo.TemperatureUnit == "fahrenheit" {
			stats.CPUTemp = (temp * 9 / 5) + 32
		} else {
			stats.CPUTemp = temp
		}
	}

	// Collect memory stats
	memUsed, memTotal, err := sc.memCollector.GetMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory stats: %w", err)
	}
	stats.MemoryUsed = memUsed
	stats.MemoryTotal = memTotal

	// Collect disk stats
	diskUsed, diskTotal, err := sc.diskCollector.GetDisk()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk stats: %w", err)
	}
	stats.DiskUsed = diskUsed
	stats.DiskTotal = diskTotal

	// Collect network interfaces
	interfaces, err := sc.netCollector.GetInterfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}
	stats.Interfaces = interfaces

	return stats, nil
}
