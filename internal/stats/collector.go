package stats

// SystemStats contains all collected system information
type SystemStats struct {
	Hostname    string
	CPUTemp     float64  // in degrees Celsius
	MemoryUsed  uint64   // in bytes
	MemoryTotal uint64   // in bytes
	DiskUsed    uint64   // in bytes
	DiskTotal   uint64   // in bytes
	Interfaces  []NetInterface
	LoadAvg1    float64 // 1-minute load average
	LoadAvg5    float64 // 5-minute load average
	LoadAvg15   float64 // 15-minute load average
	NumCPU      int     // number of logical CPUs
}

// NetInterface represents a network interface with its addresses
type NetInterface struct {
	Name      string
	IPv4Addrs []string
	IPv6Addrs []string
}

// Collector is the interface for collecting system statistics
type Collector interface {
	Collect() (*SystemStats, error)
}

// MemoryPercent returns memory usage as a percentage
func (s *SystemStats) MemoryPercent() float64 {
	if s.MemoryTotal == 0 {
		return 0
	}
	return (float64(s.MemoryUsed) / float64(s.MemoryTotal)) * 100
}

// DiskPercent returns disk usage as a percentage
func (s *SystemStats) DiskPercent() float64 {
	if s.DiskTotal == 0 {
		return 0
	}
	return (float64(s.DiskUsed) / float64(s.DiskTotal)) * 100
}

// MemoryUsedGB returns memory used in gigabytes
func (s *SystemStats) MemoryUsedGB() float64 {
	return float64(s.MemoryUsed) / (1024 * 1024 * 1024)
}

// MemoryTotalGB returns total memory in gigabytes
func (s *SystemStats) MemoryTotalGB() float64 {
	return float64(s.MemoryTotal) / (1024 * 1024 * 1024)
}

// DiskUsedGB returns disk used in gigabytes
func (s *SystemStats) DiskUsedGB() float64 {
	return float64(s.DiskUsed) / (1024 * 1024 * 1024)
}

// DiskTotalGB returns total disk in gigabytes
func (s *SystemStats) DiskTotalGB() float64 {
	return float64(s.DiskTotal) / (1024 * 1024 * 1024)
}
