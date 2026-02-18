package renderer

import (
	"fmt"
	"image"
	"image/color"

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
		// Compact all-in-one view: each segment in its own colour
		if len(layout.ContentLines) > 0 {
			y := layout.ContentLines[0]
			x := MarginLeft

			diskPct := s.DiskPercent()
			memPct := s.MemoryPercent()

			diskSeg := fmt.Sprintf("D:%.0f%%", diskPct)
			memSeg := fmt.Sprintf(" R:%.0f%%", memPct)

			if err := DrawTextColor(disp, x, y, diskSeg, MetricColor(diskPct)); err != nil {
				return err
			}
			x += MeasureText(diskSeg)

			if err := DrawTextColor(disp, x, y, memSeg, MetricColor(memPct)); err != nil {
				return err
			}
			x += MeasureText(memSeg)

			if s.CPUTemp > 0 {
				cpuSeg := fmt.Sprintf(" C:%.0fC", s.CPUTemp)
				if err := DrawTextColor(disp, x, y, cpuSeg, TempColor(s.CPUTemp)); err != nil {
					return err
				}
			}
		}
	} else if layout.Height <= 32 {
		// Small display, individual metric page: icon + coloured text
		initIcons()
		iconMaxWidth := maxWidth - IconWidth - IconGap
		var icon *image.Gray
		var text string
		var c color.NRGBA
		switch p.metricType {
		case SystemMetricDisk:
			icon = iconDisk
			text = fmt.Sprintf("%.1f/%.1fG", s.DiskUsedGB(), s.DiskTotalGB())
			c = MetricColor(s.DiskPercent())
		case SystemMetricMemory:
			icon = iconMemory
			text = fmt.Sprintf("%.1f/%.1fG", s.MemoryUsedGB(), s.MemoryTotalGB())
			c = MetricColor(s.MemoryPercent())
		case SystemMetricCPU:
			icon = iconCPU
			if s.CPUTemp > 0 {
				text = fmt.Sprintf("%.1fC", s.CPUTemp)
				c = TempColor(s.CPUTemp)
			} else {
				text = "N/A"
				c = ColorGreen
			}
		}
		text = TruncateText(text, iconMaxWidth)
		if len(layout.ContentLines) > 0 {
			if err := DrawIconTextColor(disp, MarginLeft, layout.ContentLines[0], icon, text, c); err != nil {
				return err
			}
		}
	} else {
		// Standard display: icon + coloured text for each metric
		initIcons()
		iconMaxWidth := maxWidth - IconWidth - IconGap

		type iconLine struct {
			icon  *image.Gray
			text  string
			color color.NRGBA
		}
		lines := []iconLine{
			{iconDisk, fmt.Sprintf("%.1f%% (%.1f/%.1fGB)",
				s.DiskPercent(), s.DiskUsedGB(), s.DiskTotalGB()),
				MetricColor(s.DiskPercent())},
			{iconMemory, fmt.Sprintf("%.1f%% (%.1f/%.1fGB)",
				s.MemoryPercent(), s.MemoryUsedGB(), s.MemoryTotalGB()),
				MetricColor(s.MemoryPercent())},
		}
		if s.CPUTemp > 0 {
			lines = append(lines, iconLine{iconCPU, fmt.Sprintf("%.1fC", s.CPUTemp),
				TempColor(s.CPUTemp)})
		} else {
			lines = append(lines, iconLine{iconCPU, "N/A", ColorGreen})
		}

		for i, line := range lines {
			if i >= len(layout.ContentLines) {
				break
			}
			text := TruncateText(line.text, iconMaxWidth)
			if err := DrawIconTextColor(disp, MarginLeft, layout.ContentLines[i], line.icon, text, line.color); err != nil {
				return err
			}
		}
	}

	// Show the display
	return disp.Show()
}
