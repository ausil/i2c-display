package renderer

import (
	"github.com/ausil/i2c-display/internal/config"
	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/stats"
)

// Renderer manages page rendering
type Renderer struct {
	display display.Display
	pages   []Page
	config  *config.Config
}

// NewRenderer creates a new renderer
func NewRenderer(disp display.Display, cfg *config.Config) *Renderer {
	return &Renderer{
		display: disp,
		config:  cfg,
	}
}

// BuildPages creates pages based on current statistics
func (r *Renderer) BuildPages(s *stats.SystemStats) {
	pages := make([]Page, 0)

	// Always add system page first
	pages = append(pages, NewSystemPage())

	// Add network pages based on interface count
	if len(s.Interfaces) > 0 {
		maxPerPage := r.config.Network.MaxInterfacesPerPage
		totalPages := (len(s.Interfaces) + maxPerPage - 1) / maxPerPage

		for i := 0; i < totalPages; i++ {
			pages = append(pages, NewNetworkPage(i+1, maxPerPage, len(s.Interfaces)))
		}
	}

	r.pages = pages
}

// GetPages returns the current pages
func (r *Renderer) GetPages() []Page {
	return r.pages
}

// RenderPage renders a specific page by index
func (r *Renderer) RenderPage(pageIdx int, s *stats.SystemStats) error {
	if pageIdx < 0 || pageIdx >= len(r.pages) {
		return nil // Silently ignore invalid page index
	}

	return r.pages[pageIdx].Render(r.display, s)
}

// PageCount returns the number of pages
func (r *Renderer) PageCount() int {
	return len(r.pages)
}
