package renderer

import (
	"fmt"
	"image"

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
//
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

	// Render content based on display size and metric type
	if layout.Height <= 32 && p.metricType == SystemMetricAll {
		// Compact all-in-one view: text labels only (no room for icons)
		var text string
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
		text = TruncateText(text, maxWidth)
		if len(layout.ContentLines) > 0 {
			if err := DrawText(disp, MarginLeft, layout.ContentLines[0], text); err != nil {
				return err
			}
		}
	} else if layout.Height <= 32 {
		// Small display, individual metric page: icon + text
		initIcons()
		iconMaxWidth := maxWidth - IconWidth - IconGap
		var icon *image.Gray
		var text string
		switch p.metricType {
		case SystemMetricDisk:
			icon = iconDisk
			text = fmt.Sprintf("%.1f/%.1fG", s.DiskUsedGB(), s.DiskTotalGB())
		case SystemMetricMemory:
			icon = iconMemory
			text = fmt.Sprintf("%.1f/%.1fG", s.MemoryUsedGB(), s.MemoryTotalGB())
		case SystemMetricCPU:
			icon = iconCPU
			if s.CPUTemp > 0 {
				text = fmt.Sprintf("%.1fC", s.CPUTemp)
			} else {
				text = "N/A"
			}
		}
		text = TruncateText(text, iconMaxWidth)
		if len(layout.ContentLines) > 0 {
			if err := DrawIconText(disp, MarginLeft, layout.ContentLines[0], icon, text); err != nil {
				return err
			}
		}
	} else {
		// Standard display: icon + text for each metric
		initIcons()
		iconMaxWidth := maxWidth - IconWidth - IconGap

		type iconLine struct {
			icon *image.Gray
			text string
		}
		lines := []iconLine{
			{iconDisk, fmt.Sprintf("%.1f%% (%.1f/%.1fGB)",
				s.DiskPercent(), s.DiskUsedGB(), s.DiskTotalGB())},
			{iconMemory, fmt.Sprintf("%.1f%% (%.1f/%.1fGB)",
				s.MemoryPercent(), s.MemoryUsedGB(), s.MemoryTotalGB())},
		}
		if s.CPUTemp > 0 {
			lines = append(lines, iconLine{iconCPU, fmt.Sprintf("%.1fC", s.CPUTemp)})
		} else {
			lines = append(lines, iconLine{iconCPU, "N/A"})
		}

		for i, line := range lines {
			if i >= len(layout.ContentLines) {
				break
			}
			text := TruncateText(line.text, iconMaxWidth)
			if err := DrawIconText(disp, MarginLeft, layout.ContentLines[i], line.icon, text); err != nil {
				return err
			}
		}
	}

	// Show the display
	return disp.Show()
}
