package stats

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// MemoryCollector collects memory statistics
type MemoryCollector struct {
	meminfoPath string
}

// NewMemoryCollector creates a new memory collector
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{
		meminfoPath: "/proc/meminfo",
	}
}

// NewMemoryCollectorWithPath creates a memory collector with custom path (for testing)
func NewMemoryCollectorWithPath(path string) *MemoryCollector {
	return &MemoryCollector{
		meminfoPath: path,
	}
}

// GetMemory reads memory statistics from /proc/meminfo
// Returns used and total memory in bytes
func (m *MemoryCollector) GetMemory() (uint64, uint64, error) {
	file, err := os.Open(m.meminfoPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open %s: %w", m.meminfoPath, err)
	}
	defer file.Close()

	var memTotal, memFree, memAvailable, buffers, cached uint64
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		// Values in /proc/meminfo are in kB, convert to bytes
		value *= 1024

		switch key {
		case "MemTotal":
			memTotal = value
		case "MemFree":
			memFree = value
		case "MemAvailable":
			memAvailable = value
		case "Buffers":
			buffers = value
		case "Cached":
			cached = value
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("error reading %s: %w", m.meminfoPath, err)
	}

	if memTotal == 0 {
		return 0, 0, fmt.Errorf("could not read MemTotal from %s", m.meminfoPath)
	}

	// Calculate used memory
	// Prefer MemAvailable if available (more accurate), otherwise calculate
	var memUsed uint64
	if memAvailable > 0 {
		memUsed = memTotal - memAvailable
	} else {
		// Older kernels don't have MemAvailable
		memUsed = memTotal - memFree - buffers - cached
	}

	return memUsed, memTotal, nil
}
