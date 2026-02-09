package stats

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CPUTempCollector collects CPU temperature
type CPUTempCollector struct {
	source string
}

// NewCPUTempCollector creates a new CPU temperature collector
func NewCPUTempCollector(source string) *CPUTempCollector {
	return &CPUTempCollector{
		source: source,
	}
}

// GetTemperature reads the CPU temperature from sysfs
// Returns temperature in Celsius
func (c *CPUTempCollector) GetTemperature() (float64, error) {
	data, err := os.ReadFile(c.source)
	if err != nil {
		return 0, fmt.Errorf("failed to read temperature from %s: %w", c.source, err)
	}

	// The temperature is typically in millidegrees Celsius
	tempStr := strings.TrimSpace(string(data))
	tempMilli, err := strconv.ParseInt(tempStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse temperature: %w", err)
	}

	// Convert from millidegrees to degrees
	return float64(tempMilli) / 1000.0, nil
}
