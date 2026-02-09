package stats

import (
	"fmt"
	"syscall"
)

// DiskCollector collects disk usage statistics
type DiskCollector struct {
	path string
}

// NewDiskCollector creates a new disk collector
func NewDiskCollector(path string) *DiskCollector {
	return &DiskCollector{
		path: path,
	}
}

// GetDisk reads disk usage statistics using statfs
// Returns used and total disk space in bytes
func (d *DiskCollector) GetDisk() (uint64, uint64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(d.path, &stat)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to stat filesystem at %s: %w", d.path, err)
	}

	// Total size = blocks * block size
	total := stat.Blocks * uint64(stat.Bsize)

	// Available space = available blocks * block size
	available := stat.Bavail * uint64(stat.Bsize)

	// Used space = total - available
	used := total - available

	return used, total, nil
}
