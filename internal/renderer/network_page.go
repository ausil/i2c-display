package renderer

import (
	"fmt"

	"github.com/denniskorablev/ssd1306-display/internal/display"
	"github.com/denniskorablev/ssd1306-display/internal/stats"
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
func (p *NetworkPage) Render(disp display.Display, s *stats.SystemStats) error {
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

	bounds := disp.GetBounds()
	maxWidth := bounds.Dx() - 2*MarginLeft

	// Render interfaces for this page
	yPositions := []int{Line3Y, Line4Y, Line5Y}
	interfaceCount := 0

	for i := p.interfaceStartIdx; i < p.interfaceEndIdx && i < len(s.Interfaces); i++ {
		if interfaceCount >= len(yPositions) {
			break
		}

		iface := s.Interfaces[i]
		y := yPositions[interfaceCount]

		// Determine which address to show
		var addr string
		if len(iface.IPv4Addrs) > 0 {
			addr = iface.IPv4Addrs[0]
		} else if len(iface.IPv6Addrs) > 0 {
			addr = iface.IPv6Addrs[0]
		} else {
			addr = "no addr"
		}

		text := fmt.Sprintf("%s: %s", iface.Name, addr)
		text = TruncateText(text, maxWidth)

		if err := DrawText(disp, MarginLeft, y, text); err != nil {
			return err
		}

		interfaceCount++
	}

	// Line 6: Page indicator
	if p.totalPages > 1 {
		pageIndicator := fmt.Sprintf("Page %d/%d", p.pageNum, p.totalPages)
		indicatorWidth := MeasureText(pageIndicator)
		x := bounds.Dx() - indicatorWidth - MarginRight
		if err := DrawText(disp, x, Line6Y, pageIndicator); err != nil {
			return err
		}
	}

	// Show the display
	return disp.Show()
}
