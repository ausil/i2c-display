package renderer

import (
	"fmt"

	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/stats"
)

// NetworkPage displays network interface information
type NetworkPage struct {
	pageNum            int
	maxPerPage         int
	totalPages         int
	interfaceStartIdx  int
	interfaceEndIdx    int
}

// NewNetworkPage creates a new network page
func NewNetworkPage(pageNum, maxPerPage, totalInterfaces int) *NetworkPage {
	startIdx := (pageNum - 1) * maxPerPage
	endIdx := startIdx + maxPerPage
	if endIdx > totalInterfaces {
		endIdx = totalInterfaces
	}

	totalPages := (totalInterfaces + maxPerPage - 1) / maxPerPage

	return &NetworkPage{
		pageNum:           pageNum,
		maxPerPage:        maxPerPage,
		totalPages:        totalPages,
		interfaceStartIdx: startIdx,
		interfaceEndIdx:   endIdx,
	}
}

// Title returns the page title
func (p *NetworkPage) Title() string {
	return fmt.Sprintf("Network %d/%d", p.pageNum, p.totalPages)
}

// Render draws the network page
//nolint:gocyclo // rendering logic naturally has many conditional branches for different display sizes
func (p *NetworkPage) Render(disp display.Display, s *stats.SystemStats) error {
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

	// Render interfaces for this page
	interfaceCount := 0

	for i := p.interfaceStartIdx; i < p.interfaceEndIdx && i < len(s.Interfaces); i++ {
		if interfaceCount >= len(layout.ContentLines) {
			break
		}

		iface := s.Interfaces[i]
		y := layout.ContentLines[interfaceCount]

		// Determine which address to show
		var addr string
		if len(iface.IPv4Addrs) > 0 {
			addr = iface.IPv4Addrs[0]
		} else if len(iface.IPv6Addrs) > 0 {
			addr = iface.IPv6Addrs[0]
		} else {
			addr = "no addr"
		}

		// Format based on display size
		var text string
		if layout.Height <= 32 {
			// Compact format for small displays: "name:IP"
			// Use shorter separator to save space
			text = fmt.Sprintf("%s:%s", iface.Name, addr)
		} else {
			// Standard format: "interface: IP"
			text = fmt.Sprintf("%s: %s", iface.Name, addr)
		}

		text = TruncateText(text, maxWidth)
		if err := DrawText(disp, MarginLeft, y, text); err != nil {
			return err
		}

		interfaceCount++
	}

	// Footer: Page indicator (if space available and multiple pages)
	if p.totalPages > 1 && layout.FooterY >= 0 {
		pageIndicator := fmt.Sprintf("Page %d/%d", p.pageNum, p.totalPages)
		indicatorWidth := MeasureText(pageIndicator)
		x := bounds.Dx() - indicatorWidth - MarginRight
		if err := DrawText(disp, x, layout.FooterY, pageIndicator); err != nil {
			return err
		}
	}

	// Show the display
	return disp.Show()
}
