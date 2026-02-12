package renderer

import (
	"sync"

	"github.com/ausil/i2c-display/internal/config"
	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/stats"
)

// Renderer manages page rendering
type Renderer struct {
	display display.Display
	pages   []Page
	mu      sync.RWMutex // Protects pages slice
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

	// For small displays (128x32), create separate pages for each metric
	bounds := r.display.GetBounds()
	if bounds.Dy() <= 32 {
		// Add individual metric pages for better readability
		pages = append(pages, NewSystemPageForMetric(SystemMetricDisk), NewSystemPageForMetric(SystemMetricMemory))
		if s.CPUTemp > 0 {
			pages = append(pages, NewSystemPageForMetric(SystemMetricCPU))
		}
	} else {
		// Standard displays show all system info on one page
		pages = append(pages, NewSystemPage())
	}

	// Add network pages based on interface count
	if len(s.Interfaces) > 0 {
		maxPerPage := r.config.Network.MaxInterfacesPerPage
		totalPages := (len(s.Interfaces) + maxPerPage - 1) / maxPerPage

		for i := 0; i < totalPages; i++ {
			pages = append(pages, NewNetworkPage(i+1, maxPerPage, len(s.Interfaces)))
		}
	}

	r.mu.Lock()
	r.pages = pages
	r.mu.Unlock()
}

// GetPages returns the current pages
func (r *Renderer) GetPages() []Page {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// Return a copy to prevent external modification
	pagesCopy := make([]Page, len(r.pages))
	copy(pagesCopy, r.pages)
	return pagesCopy
}

// RenderPage renders a specific page by index
func (r *Renderer) RenderPage(pageIdx int, s *stats.SystemStats) error {
	r.mu.RLock()
	if pageIdx < 0 || pageIdx >= len(r.pages) {
		r.mu.RUnlock()
		return nil // Silently ignore invalid page index
	}
	page := r.pages[pageIdx]
	r.mu.RUnlock()

	return page.Render(r.display, s)
}

// PageCount returns the number of pages
func (r *Renderer) PageCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.pages)
}
