package stats

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const defaultLoadAvgPath = "/proc/loadavg"

// LoadAvgCollector collects system load averages
type LoadAvgCollector struct {
	path string
}

// NewLoadAvgCollector creates a new load average collector
func NewLoadAvgCollector() *LoadAvgCollector {
	return &LoadAvgCollector{path: defaultLoadAvgPath}
}

// NewLoadAvgCollectorWithPath creates a collector reading from a custom path (for testing)
func NewLoadAvgCollectorWithPath(path string) *LoadAvgCollector {
	return &LoadAvgCollector{path: path}
}

// GetLoadAvg reads /proc/loadavg and returns the 1m, 5m, and 15m load averages
func (c *LoadAvgCollector) GetLoadAvg() (avg1, avg5, avg15 float64, err error) {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to read load average from %s: %w", c.path, err)
	}

	fields := strings.Fields(strings.TrimSpace(string(data)))
	if len(fields) < 3 {
		return 0, 0, 0, fmt.Errorf("unexpected loadavg format: %q", string(data))
	}

	avg1, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse 1m load average %q: %w", fields[0], err)
	}

	avg5, err = strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse 5m load average %q: %w", fields[1], err)
	}

	avg15, err = strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse 15m load average %q: %w", fields[2], err)
	}

	return avg1, avg5, avg15, nil
}
