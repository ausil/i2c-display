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

	// Line 1: Hostname (centered)
	if err := DrawTextCentered(disp, Line1Y, s.Hostname); err != nil {
		return err
	}

	// Line 2: Separator
	if err := DrawLine(disp, Line2Y); err != nil {
		return err
	}

	// Line 3: Disk usage
	diskText := fmt.Sprintf("Disk: %.1f%% (%.1f/%.1fGB)",
		s.DiskPercent(),
		s.DiskUsedGB(),
		s.DiskTotalGB())
	bounds := disp.GetBounds()
	diskText = TruncateText(diskText, bounds.Dx()-2*MarginLeft)
	if err := DrawText(disp, MarginLeft, Line3Y, diskText); err != nil {
		return err
	}

	// Line 4: RAM usage
	ramText := fmt.Sprintf("RAM: %.1f%% (%.1f/%.1fGB)",
		s.MemoryPercent(),
		s.MemoryUsedGB(),
		s.MemoryTotalGB())
	ramText = TruncateText(ramText, bounds.Dx()-2*MarginLeft)
	if err := DrawText(disp, MarginLeft, Line4Y, ramText); err != nil {
		return err
	}

	// Line 5: CPU temperature
	var cpuText string
	if s.CPUTemp > 0 {
		cpuText = fmt.Sprintf("CPU: %.1fC", s.CPUTemp)
	} else {
		cpuText = "CPU: N/A"
	}
	if err := DrawText(disp, MarginLeft, Line5Y, cpuText); err != nil {
		return err
	}

	// Show the display
	return disp.Show()
}
