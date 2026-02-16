package renderer

import (
	"fmt"

	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/stats"
)

// SystemMetricType represents the type of metric to display
type SystemMetricType int

const (
	SystemMetricAll SystemMetricType = iota
	SystemMetricDisk
	SystemMetricMemory
	SystemMetricCPU
)

// SystemPage displays system statistics (disk, RAM, CPU temp)
type SystemPage struct {
	metricType SystemMetricType
}

// NewSystemPage creates a new system stats page showing all metrics
func NewSystemPage() *SystemPage {
	return &SystemPage{metricType: SystemMetricAll}
}

// NewSystemPageForMetric creates a system page for a specific metric
func NewSystemPageForMetric(metricType SystemMetricType) *SystemPage {
	return &SystemPage{metricType: metricType}
}

// Title returns the page title
func (p *SystemPage) Title() string {
	switch p.metricType {
	case SystemMetricDisk:
		return "Disk"
	case SystemMetricMemory:
		return "Memory"
	case SystemMetricCPU:
		return "CPU"
	default:
		return "System"
	}
}

// Render draws the system stats page
//nolint:gocyclo // rendering logic naturally has many conditional branches for different display sizes
func (p *SystemPage) Render(disp display.Display, s *stats.SystemStats) error {
	// Clear display
	if err := disp.Clear(); err != nil {
		return err
	}

	// Create adaptive layout
	bounds := disp.GetBounds()
	layout := NewLayout(bounds)
	maxWidth := bounds.Dx() - 2*MarginLeft

	// Optional: Hostname header (green on colour displays)
	if layout.ShowHeader {
		if err := DrawTextCenteredColor(disp, layout.HeaderY, s.Hostname, ColorGreen); err != nil {
			return err
		}
	}

	// Optional: Separator
	if layout.ShowSeparator {
		if err := DrawLine(disp, layout.SeparatorY); err != nil {
			return err
		}
	}

	// Build content lines based on available space
	contentLines := make([]string, 0, layout.MaxContentLines)

	// For small displays (128x32), show one metric at a time
	if layout.Height <= 32 {
		var text string
		switch p.metricType {
		case SystemMetricDisk:
			text = fmt.Sprintf("Disk: %.1f/%.1fG", s.DiskUsedGB(), s.DiskTotalGB())
		case SystemMetricMemory:
			text = fmt.Sprintf("Mem: %.1f/%.1fG", s.MemoryUsedGB(), s.MemoryTotalGB())
		case SystemMetricCPU:
			if s.CPUTemp > 0 {
				text = fmt.Sprintf("CPU Temp: %.1fC", s.CPUTemp)
			} else {
				text = "CPU Temp: N/A"
			}
		default:
			// Fallback to compact all-in-one view
			if s.CPUTemp > 0 {
				text = fmt.Sprintf("D:%.0f%% R:%.0f%% C:%.0fC",
					s.DiskPercent(),
					s.MemoryPercent(),
					s.CPUTemp)
			} else {
				text = fmt.Sprintf("D:%.0f%% R:%.0f%%",
					s.DiskPercent(),
					s.MemoryPercent())
			}
		}
		contentLines = append(contentLines, text)
	} else {
		// Standard display - show full info
		// Disk usage
		diskText := fmt.Sprintf("Disk: %.1f%% (%.1f/%.1fGB)",
			s.DiskPercent(),
			s.DiskUsedGB(),
			s.DiskTotalGB())
		contentLines = append(contentLines, diskText)

		// RAM usage
		ramText := fmt.Sprintf("RAM: %.1f%% (%.1f/%.1fGB)",
			s.MemoryPercent(),
			s.MemoryUsedGB(),
			s.MemoryTotalGB())
		contentLines = append(contentLines, ramText)

		// CPU temperature
		var cpuText string
		if s.CPUTemp > 0 {
			cpuText = fmt.Sprintf("CPU: %.1fC", s.CPUTemp)
		} else {
			cpuText = "CPU: N/A"
		}
		contentLines = append(contentLines, cpuText)
	}

	// Render content lines
	for i, text := range contentLines {
		if i >= len(layout.ContentLines) {
			break // Don't exceed available lines
		}
		text = TruncateText(text, maxWidth)
		if err := DrawText(disp, MarginLeft, layout.ContentLines[i], text); err != nil {
			return err
		}
	}

	// Show the display
	return disp.Show()
}
