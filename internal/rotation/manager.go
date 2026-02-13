package rotation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ausil/i2c-display/internal/config"
	"github.com/ausil/i2c-display/internal/logger"
	"github.com/ausil/i2c-display/internal/metrics"
	"github.com/ausil/i2c-display/internal/renderer"
	"github.com/ausil/i2c-display/internal/stats"
)

// Manager handles page rotation and refresh
type Manager struct {
	config               *config.Config
	collector            *stats.SystemCollector
	renderer             *renderer.Renderer
	log                  *logger.Logger
	metricsCollector     *metrics.Collector // optional, nil if metrics disabled
	currentPage          int
	lastInterfaceCount   int
	mu                   sync.Mutex // Protects currentPage and lastInterfaceCount
	stopOnce             sync.Once
	rotationTicker       *time.Ticker
	refreshTicker        *time.Ticker
	stopChan             chan struct{}
	stoppedChan          chan struct{}
}

// SetMetrics attaches a metrics collector to the manager.
// Must be called before Start.
func (m *Manager) SetMetrics(c *metrics.Collector) {
	m.metricsCollector = c
}

// NewManager creates a new rotation manager
func NewManager(cfg *config.Config, collector *stats.SystemCollector, rend *renderer.Renderer) *Manager {
	return &Manager{
		config:             cfg,
		collector:          collector,
		renderer:           rend,
		log:                logger.Global(),
		currentPage:        0,
		lastInterfaceCount: -1, // -1 forces a BuildPages on the first refresh
		stopChan:           make(chan struct{}),
		stoppedChan:        make(chan struct{}),
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
	defer func() {
		if r := recover(); r != nil {
			m.log.Errorf("PANIC in rotation manager: %v", r)
		}
		close(m.stoppedChan)
	}()
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
				m.log.ErrorWithErr(err, "refresh error")
			}
		}
	}
}

// refreshCurrentPage collects new stats and re-renders the current page
func (m *Manager) refreshCurrentPage() error {
	// Collect current stats
	systemStats, err := m.collector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect stats: %w", err)
	}

	// Only rebuild pages when the interface count changes to avoid unnecessary work
	m.mu.Lock()
	interfaceCountChanged := len(systemStats.Interfaces) != m.lastInterfaceCount
	if interfaceCountChanged {
		m.lastInterfaceCount = len(systemStats.Interfaces)
	}
	m.mu.Unlock()

	if interfaceCountChanged {
		m.renderer.BuildPages(systemStats)
	}

	// Ensure current page is valid after any rebuild
	m.mu.Lock()
	if m.currentPage >= m.renderer.PageCount() {
		m.currentPage = 0
	}
	pageIdx := m.currentPage
	m.mu.Unlock()

	// Render current page
	start := time.Now()
	err = m.renderer.RenderPage(pageIdx, systemStats)
	if m.metricsCollector != nil {
		m.metricsCollector.RecordDisplayRefresh(err == nil, time.Since(start), "system")
		m.metricsCollector.UpdateSystemMetrics(
			systemStats.CPUTemp,
			systemStats.MemoryPercent(),
			systemStats.DiskPercent(),
			len(systemStats.Interfaces),
		)
	}
	return err
}

// rotatePage advances to the next page
func (m *Manager) rotatePage() {
	m.mu.Lock()
	m.currentPage++
	if m.currentPage >= m.renderer.PageCount() {
		m.currentPage = 0
	}
	page := m.currentPage
	m.mu.Unlock()

	if m.metricsCollector != nil {
		m.metricsCollector.RecordPageRotation(page)
	}
	// Refresh will happen on next refresh tick
}

// Stop stops the rotation manager gracefully
func (m *Manager) Stop() {
	m.stopOnce.Do(func() {
		close(m.stopChan)
	})

	// Wait for goroutine to stop with timeout to prevent deadlock
	select {
	case <-m.stoppedChan:
		// Normal shutdown
	case <-time.After(5 * time.Second):
		m.log.Warn("rotation manager stop timed out")
	}
}

// CurrentPage returns the current page index
func (m *Manager) CurrentPage() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentPage
}
