package screensaver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/logger"
)

// Mode represents the screen saver mode
type Mode string

const (
	// ModeOff - screen saver disabled
	ModeOff Mode = "off"
	// ModeDim - dim the display after inactivity
	ModeDim Mode = "dim"
	// ModeBlank - turn off display after inactivity
	ModeBlank Mode = "blank"
)

// Config holds screen saver configuration
type Config struct {
	Enabled          bool          `json:"enabled"`
	Mode             Mode          `json:"mode"`              // "off", "dim", or "blank"
	IdleTimeout      time.Duration `json:"idle_timeout"`      // Time before activation
	DimBrightness    uint8         `json:"dim_brightness"`    // Brightness when dimmed (0-255)
	NormalBrightness uint8         `json:"normal_brightness"` // Normal operating brightness
}

// ScreenSaver manages display power saving
type ScreenSaver struct {
	cfg        Config
	disp       display.Display
	log        *logger.Logger
	mu         sync.RWMutex
	lastActive time.Time
	isActive   bool // true if screen saver is currently active
	ticker     *time.Ticker
	stopChan   chan struct{}
}

// New creates a new screen saver
func New(cfg Config, disp display.Display, log *logger.Logger) *ScreenSaver {
	return &ScreenSaver{
		cfg:        cfg,
		disp:       disp,
		log:        log,
		lastActive: time.Now(),
		isActive:   false,
		stopChan:   make(chan struct{}),
	}
}

// Start starts the screen saver monitor
func (s *ScreenSaver) Start(ctx context.Context) error {
	if !s.cfg.Enabled {
		s.log.Debug("Screen saver disabled")
		return nil
	}

	s.log.With().
		Str("mode", string(s.cfg.Mode)).
		Str("timeout", s.cfg.IdleTimeout.String()).
		Logger().Info("Starting screen saver")

	// Set initial brightness
	if err := s.disp.SetBrightness(s.cfg.NormalBrightness); err != nil {
		s.log.ErrorWithErr(err, "Failed to set initial brightness")
	}

	// Check every 10 seconds
	s.ticker = time.NewTicker(10 * time.Second)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Log panic but don't crash the application
				s.log.With().Str("panic", fmt.Sprintf("%v", r)).Logger().Error("PANIC in screen saver")
			}
		}()

		for {
			select {
			case <-ctx.Done():
				s.ticker.Stop()
				return
			case <-s.stopChan:
				s.ticker.Stop()
				return
			case <-s.ticker.C:
				s.check()
			}
		}
	}()

	return nil
}

// Stop stops the screen saver
func (s *ScreenSaver) Stop() {
	if !s.cfg.Enabled {
		return
	}

	close(s.stopChan)
	s.log.Debug("Screen saver stopped")
}

// check checks if screen saver should activate or deactivate
func (s *ScreenSaver) check() {
	s.mu.Lock()
	idle := time.Since(s.lastActive)
	shouldActivate := idle >= s.cfg.IdleTimeout && !s.isActive
	shouldDeactivate := idle < s.cfg.IdleTimeout && s.isActive
	s.mu.Unlock()

	// Call activate/deactivate without holding the lock
	if shouldActivate {
		s.activate()
	} else if shouldDeactivate {
		s.deactivate()
	}
}

// activate activates the screen saver
func (s *ScreenSaver) activate() {
	s.log.With().Str("mode", string(s.cfg.Mode)).Logger().Info("Activating screen saver")

	// Perform display operations without holding the lock
	var err error
	switch s.cfg.Mode {
	case ModeDim:
		err = s.disp.SetBrightness(s.cfg.DimBrightness)
	case ModeBlank:
		err = s.disp.SetBrightness(0)
	}

	if err != nil {
		s.log.ErrorWithErr(err, "Failed to activate screen saver")
		return
	}

	// Only set isActive flag if brightness change succeeded
	s.mu.Lock()
	s.isActive = true
	s.mu.Unlock()
}

// deactivate deactivates the screen saver
func (s *ScreenSaver) deactivate() {
	s.log.Debug("Deactivating screen saver")

	// Perform display operation without holding the lock
	if err := s.disp.SetBrightness(s.cfg.NormalBrightness); err != nil {
		s.log.ErrorWithErr(err, "Failed to restore brightness")
		return
	}

	// Only clear isActive flag if brightness change succeeded
	s.mu.Lock()
	s.isActive = false
	s.mu.Unlock()
}

// ResetActivity resets the idle timer (call when user activity detected)
func (s *ScreenSaver) ResetActivity() {
	if !s.cfg.Enabled {
		return
	}

	s.mu.Lock()
	wasActive := s.isActive
	s.lastActive = time.Now()
	s.mu.Unlock()

	// If screen saver was active, deactivate immediately (without holding lock)
	if wasActive {
		s.deactivate()
	}
}

// IsActive returns whether the screen saver is currently active
func (s *ScreenSaver) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isActive
}

// Config returns the current screen saver configuration
func (s *ScreenSaver) Config() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

// UpdateConfig updates the screen saver configuration
func (s *ScreenSaver) UpdateConfig(cfg Config) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldEnabled := s.cfg.Enabled
	s.cfg = cfg

	// If enabling or disabling, handle accordingly
	if !oldEnabled && cfg.Enabled {
		s.lastActive = time.Now()
		s.isActive = false
	} else if oldEnabled && !cfg.Enabled {
		if s.isActive {
			s.deactivate()
		}
	}
}
