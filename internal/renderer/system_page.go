package renderer

import (
	"fmt"

	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/stats"
)

// SystemPage displays system statistics (disk, RAM, CPU temp)
type SystemPage struct{}

// NewSystemPage creates a new system stats page
func NewSystemPage() *SystemPage {
	return &SystemPage{}
}

// Title returns the page title
func (p *SystemPage) Title() string {
	return "System"
}

// Render draws the system stats page
func (p *SystemPage) Render(disp display.Display, s *stats.SystemStats) error {
	// Clear display
	if err := disp.Clear(); err != nil {
		return err
	}

	// Create adaptive layout
	bounds := disp.GetBounds()
	layout := NewLayout(bounds)
	maxWidth := bounds.Dx() - 2*MarginLeft

	// Optional: Hostname header
	if layout.ShowHeader {
		if err := DrawTextCentered(disp, layout.HeaderY, s.Hostname); err != nil {
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

	// For small displays (128x32), show compact info
	if layout.Height <= 32 {
		// Line 1: Disk and RAM usage (compact)
		diskText := fmt.Sprintf("D:%.0f%% R:%.0f%%",
			s.DiskPercent(),
			s.MemoryPercent())
		contentLines = append(contentLines, diskText)
		// Line 2: Temperature (if available)
		if s.CPUTemp > 0 {
			cpuText := fmt.Sprintf("CPU:%.0fC", s.CPUTemp)
			contentLines = append(contentLines, cpuText)
		}
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
