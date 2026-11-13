package renderer

import (
	"fmt"

	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/stats"
)

// NetworkPage displays network interface information
type NetworkPage struct {
	pageNum           int
	maxPerPage        int
	totalPages        int
	interfaceStartIdx int
	interfaceEndIdx   int
	lines             int // configured line count (0=auto, 2=default, 4=compact)
}

// NewNetworkPage creates a new network page
func NewNetworkPage(pageNum, maxPerPage, totalInterfaces, lines int) *NetworkPage {
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
		lines:             lines,
	}
}

// networkLine formats one interface row. Small displays drop the space after
// the colon because every character costs 5-7 pixels on a 128-wide panel.
func networkLine(name, addr string, compact bool) string {
	if compact {
		return fmt.Sprintf("%s:%s", name, addr)
	}
	return fmt.Sprintf("%s: %s", name, addr)
}

// Title returns the page title
func (p *NetworkPage) Title() string {
	return fmt.Sprintf("Network %d/%d", p.pageNum, p.totalPages)
}

// Render draws the network page
//
//nolint:gocyclo // rendering logic naturally has many conditional branches for different display sizes
func (p *NetworkPage) Render(disp display.Display, s *stats.SystemStats) error {
	// Clear display
	if err := disp.Clear(); err != nil {
		return err
	}

	// Create adaptive layout
	bounds := disp.GetBounds()
	layout := NewLayout(bounds, p.lines)
	maxWidth := bounds.Dx() - 2*MarginLeft

	// Optional: Hostname header (green on colour displays)
	if layout.ShowHeader {
		if err := DrawTextCenteredColorScaled(disp, layout.HeaderY, s.Hostname, ColorGreen, layout.TextScale); err != nil {
			return err
		}
	}

	// Optional: Separator
	if layout.ShowSeparator {
		if err := DrawLine(disp, layout.SeparatorY); err != nil {
			return err
		}
	}

	// Render interfaces for this page. Lines are composed first so the widest
	// one decides the font scale for the whole page; a page mixing two font
	// sizes reads as a rendering fault.
	type pendingLine struct {
		y    int
		text string
	}
	var pending []pendingLine

	for i := p.interfaceStartIdx; i < p.interfaceEndIdx && i < len(s.Interfaces); i++ {
		if len(pending) >= len(layout.ContentLines) {
			break
		}

		iface := s.Interfaces[i]

		// Determine which address to show
		var addr string
		if len(iface.IPv4Addrs) > 0 {
			addr = iface.IPv4Addrs[0]
		} else if len(iface.IPv6Addrs) > 0 {
			addr = iface.IPv6Addrs[0]
		} else {
			addr = "no addr"
		}

		pending = append(pending, pendingLine{
			y:    layout.ContentLines[len(pending)],
			text: networkLine(iface.Name, addr, layout.Height <= 32),
		})
	}

	texts := make([]string, len(pending))
	for i, line := range pending {
		texts[i] = line.text
	}
	scale := fitScale(texts, maxWidth, layout.TextScale)

	for _, line := range pending {
		if err := DrawTextColorScaled(disp, MarginLeft, line.y, truncateAt(line.text, maxWidth, scale), ColorGreen, scale); err != nil {
			return err
		}
	}

	// Footer: Page indicator (if space available and multiple pages)
	if p.totalPages > 1 && layout.FooterY >= 0 {
		pageIndicator := fmt.Sprintf("Page %d/%d", p.pageNum, p.totalPages)
		indicatorWidth := MeasureText(pageIndicator)
		x := bounds.Dx() - indicatorWidth - MarginRight
		if err := DrawTextColorScaled(disp, x, layout.FooterY, pageIndicator, ColorGreen, layout.TextScale); err != nil {
			return err
		}
	}

	// Show the display
	return disp.Show()
}
