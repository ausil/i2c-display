package rotation

import (
	"context"
	"fmt"
	"time"

	"github.com/ausil/i2c-display/internal/config"
	"github.com/ausil/i2c-display/internal/renderer"
	"github.com/ausil/i2c-display/internal/stats"
)

// Manager handles page rotation and refresh
type Manager struct {
	config          *config.Config
	collector       *stats.SystemCollector
	renderer        *renderer.Renderer
	currentPage     int
	rotationTicker  *time.Ticker
	refreshTicker   *time.Ticker
	stopChan        chan struct{}
	stoppedChan     chan struct{}
}

// NewManager creates a new rotation manager
func NewManager(cfg *config.Config, collector *stats.SystemCollector, rend *renderer.Renderer) *Manager {
	return &Manager{
		config:      cfg,
		collector:   collector,
		renderer:    rend,
		currentPage: 0,
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}
}

// Start begins the rotation and refresh loops
func (m *Manager) Start(ctx context.Context) error {
	// Get intervals from config
	rotationInterval, err := m.config.Pages.GetRotationInterval()
	if err != nil {
		return fmt.Errorf("invalid rotation interval: %w", err)
	}

	refreshInterval, err := m.config.Pages.GetRefreshInterval()
	if err != nil {
		return fmt.Errorf("invalid refresh interval: %w", err)
	}

	// Create tickers
	m.rotationTicker = time.NewTicker(rotationInterval)
	m.refreshTicker = time.NewTicker(refreshInterval)

	// Initial render
	if err := m.refreshCurrentPage(); err != nil {
		return fmt.Errorf("initial render failed: %w", err)
	}

	// Start rotation loop
	go m.run(ctx)

	return nil
}

// run is the main rotation loop
func (m *Manager) run(ctx context.Context) {
	defer close(m.stoppedChan)
	defer m.rotationTicker.Stop()
	defer m.refreshTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopChan:
			return
		case <-m.rotationTicker.C:
			m.rotatePage()
		case <-m.refreshTicker.C:
			if err := m.refreshCurrentPage(); err != nil {
				// Log error but continue
				fmt.Printf("refresh error: %v\n", err)
			}
		}
	}
}

// refreshCurrentPage collects new stats and re-renders the current page
func (m *Manager) refreshCurrentPage() error {
	// Collect current stats
	stats, err := m.collector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect stats: %w", err)
	}

	// Rebuild pages in case interface count changed
	m.renderer.BuildPages(stats)

	// Ensure current page is valid
	if m.currentPage >= m.renderer.PageCount() {
		m.currentPage = 0
	}

	// Render current page
	return m.renderer.RenderPage(m.currentPage, stats)
}

// rotatePage advances to the next page
func (m *Manager) rotatePage() {
	m.currentPage++
	if m.currentPage >= m.renderer.PageCount() {
		m.currentPage = 0
	}

	// Refresh will happen on next refresh tick
}

// Stop stops the rotation manager gracefully
func (m *Manager) Stop() {
	close(m.stopChan)
	<-m.stoppedChan
}

// CurrentPage returns the current page index
func (m *Manager) CurrentPage() int {
	return m.currentPage
}
