package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ausil/i2c-display/internal/config"
	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/logger"
	"github.com/ausil/i2c-display/internal/metrics"
	"github.com/ausil/i2c-display/internal/renderer"
	"github.com/ausil/i2c-display/internal/rotation"
	"github.com/ausil/i2c-display/internal/screensaver"
	"github.com/ausil/i2c-display/internal/stats"
)

//nolint:funlen,gocyclo // main function naturally has many statements for initialization
func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to configuration file")
	useMock := flag.Bool("mock", false, "Use mock display (for testing without hardware)")
	validateConfig := flag.Bool("validate-config", false, "Validate configuration and exit")
	testDisplay := flag.Bool("test-display", false, "Run display hardware test pattern and exit")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadWithPriority(*configPath)
	if err != nil {
		// Use default logger before config is loaded
		log := logger.NewDefault()
		log.FatalWithErr(err, "Failed to load configuration")
	}

	// If validate-config flag is set, validate and exit
	if *validateConfig {
		log := logger.NewDefault()
		log.Info("Validating configuration...")
		if err := cfg.Validate(); err != nil {
			log.ErrorWithErr(err, "Configuration validation failed")
			os.Exit(1)
		}
		log.Info("Configuration is valid")
		os.Exit(0)
	}

	// Set up logging from config
	log := logger.New(logger.Config{
		Level:  cfg.Logging.Level,
		Output: cfg.Logging.Output,
		JSON:   cfg.Logging.JSON,
	})
	logger.SetGlobalLogger(log)

	log.Info("I2C Display Service starting...")
	log.With().Str("type", cfg.Display.Type).Logger().Info("Display configuration loaded")
	log.With().Str("mode", cfg.SystemInfo.HostnameDisplay).Logger().Info("Hostname display mode configured")

	// Create display
	var disp display.Display
	if *useMock {
		log.Info("Using mock display (no hardware)")
		disp = display.NewMockDisplay(cfg.Display.Width, cfg.Display.Height)
	} else {
		log.With().
			Str("type", cfg.Display.Type).
			Str("bus", cfg.Display.I2CBus).
			Str("address", cfg.Display.I2CAddress).
			Logger().Info("Initializing display hardware")
		hardwareDisp, err := display.NewDisplay(&cfg.Display)
		if err != nil {
			log.ErrorWithErr(err, "Failed to initialize hardware display")
			log.Warn("Falling back to mock display")
			disp = display.NewMockDisplay(cfg.Display.Width, cfg.Display.Height)
		} else {
			disp = hardwareDisp
		}
	}

	// Initialize display
	if err := disp.Init(); err != nil {
		log.FatalWithErr(err, "Failed to initialize display")
	}
	defer func() {
		log.Info("Closing display...")
		if err := disp.Close(); err != nil {
			log.ErrorWithErr(err, "Error closing display")
		}
	}()

	// Run hardware test pattern if requested
	if *testDisplay {
		log.Info("Running display test pattern...")
		if err := runDisplayTest(disp, log); err != nil {
			log.FatalWithErr(err, "Display test failed")
		}
		log.Info("Display test complete")
		return
	}

	// Create stats collector
	collector, err := stats.NewSystemCollector(cfg)
	if err != nil {
		log.FatalWithErr(err, "Failed to create stats collector")
	}

	// Create renderer
	rend := renderer.NewRenderer(disp, cfg)

	// Collect initial stats to build pages
	initialStats, err := collector.Collect()
	if err != nil {
		log.FatalWithErr(err, "Failed to collect initial stats")
	}
	rend.BuildPages(initialStats)

	log.With().Int("count", rend.PageCount()).Logger().Info("Pages built successfully")

	// Create rotation manager
	mgr := rotation.NewManager(cfg, collector, rend)

	// Create and attach metrics collector
	metricsCollector := metrics.New(log)
	mgr.SetMetrics(metricsCollector)

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start metrics server if enabled
	metricsServer, err := metrics.StartMetricsServer(metrics.Config{
		Enabled: cfg.Metrics.Enabled,
		Address: cfg.Metrics.Address,
	}, metricsCollector, log)
	if err != nil {
		log.ErrorWithErr(err, "Failed to start metrics server")
	}

	// Create and start screensaver
	ss, err := newScreenSaver(cfg, disp, log)
	if err != nil {
		log.FatalWithErr(err, "Invalid screensaver configuration")
	}
	if err := ss.Start(ctx); err != nil {
		log.ErrorWithErr(err, "Failed to start screensaver")
	}
	defer ss.Stop()

	// Start rotation manager
	if err := mgr.Start(ctx); err != nil {
		log.FatalWithErr(err, "Failed to start rotation manager")
	}

	log.Info("Display service running. Press Ctrl+C to stop.")

	// Wait for interrupt signal or SIGHUP for reload
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	for {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP:
			log.Info("Received SIGHUP, reloading configuration...")
			newCfg, err := config.LoadWithPriority(*configPath)
			if err != nil {
				log.ErrorWithErr(err, "Failed to reload configuration, keeping current config")
				continue
			}
			if err := newCfg.Validate(); err != nil {
				log.ErrorWithErr(err, "New configuration invalid, keeping current config")
				continue
			}
			// Warn if display hardware config changed — requires a restart
			if newCfg.Display != cfg.Display {
				log.Warn("Display configuration changed — restart required for changes to take effect")
			}
			// Update logging if changed
			if newCfg.Logging != cfg.Logging {
				log = logger.New(logger.Config{
					Level:  newCfg.Logging.Level,
					Output: newCfg.Logging.Output,
					JSON:   newCfg.Logging.JSON,
				})
				logger.SetGlobalLogger(log)
				log.Info("Logging configuration updated")
			}
			// Update screensaver config
			newSS, ssErr := newScreenSaver(newCfg, disp, log)
			if ssErr != nil {
				log.ErrorWithErr(ssErr, "Invalid screensaver configuration, keeping current")
			} else {
				ss.UpdateConfig(newSS.Config())
			}
			cfg = newCfg
			log.Info("Configuration reloaded successfully")
			continue

		case syscall.SIGINT, syscall.SIGTERM:
			log.With().Str("signal", sig.String()).Logger().Info("Received shutdown signal")
			goto shutdown
		}
	}

shutdown:

	// Cancel context to stop rotation manager and screensaver
	cancel()

	// Stop manager gracefully
	mgr.Stop()

	// Stop metrics server if running
	if metricsServer != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := metricsServer.Stop(shutdownCtx); err != nil {
			log.ErrorWithErr(err, "Error stopping metrics server")
		}
	}

	log.Info("Shutdown complete")
}

// runDisplayTest draws a sequence of test patterns to verify display hardware.
// Each step pauses so the result can be inspected visually.
//
// Pass: solid white fill → border rectangle → cross-hairs → text → clear.
//
//nolint:gocyclo // test sequence naturally has many steps
func runDisplayTest(disp display.Display, log *logger.Logger) error {
	bounds := disp.GetBounds()
	w := bounds.Dx()
	h := bounds.Dy()

	steps := []struct {
		name string
		fn   func() error
	}{
		{
			// Step 1 — solid white: verifies the full display area is addressed.
			// If only part of the screen lights up the window/offset is wrong.
			name: "solid white fill",
			fn: func() error {
				for y := 0; y < h; y++ {
					for x := 0; x < w; x++ {
						if err := disp.DrawPixel(x, y, true); err != nil {
							return err
						}
					}
				}
				return disp.Show()
			},
		},
		{
			// Step 2 — border: verifies all four edges reach the display corners.
			name: "border rectangle",
			fn: func() error {
				if err := disp.Clear(); err != nil {
					return err
				}
				if err := disp.DrawRect(0, 0, w, h, false); err != nil {
					return err
				}
				return disp.Show()
			},
		},
		{
			// Step 3 — cross-hairs: verifies centre coordinates and axis directions.
			name: "cross-hairs",
			fn: func() error {
				if err := disp.Clear(); err != nil {
					return err
				}
				// Horizontal centre line
				if err := disp.DrawLine(0, h/2, w); err != nil {
					return err
				}
				// Vertical centre line (pixel by pixel)
				for y := 0; y < h; y++ {
					if err := disp.DrawPixel(w/2, y, true); err != nil {
						return err
					}
				}
				return disp.Show()
			},
		},
		{
			// Step 4 — text: verifies the rendering pipeline end-to-end.
			// Text appears top-left; if it is mirrored/upside-down the
			// rotation or MADCTL value needs adjusting.
			name: "text rendering",
			fn: func() error {
				if err := disp.Clear(); err != nil {
					return err
				}
				if err := disp.DrawText(2, 2, "DISPLAY OK", display.FontSmall); err != nil {
					return err
				}
				size := fmt.Sprintf("%dx%d", w, h)
				if err := disp.DrawText(2, 14, size, display.FontSmall); err != nil {
					return err
				}
				return disp.Show()
			},
		},
		{
			// Step 5 — clear: leave the display blank.
			name: "clear",
			fn: func() error {
				if err := disp.Clear(); err != nil {
					return err
				}
				return disp.Show()
			},
		},
	}

	for i, step := range steps {
		log.With().Int("step", i+1).Str("name", step.name).Logger().Info("Test step")
		if err := step.fn(); err != nil {
			return fmt.Errorf("step %d (%s): %w", i+1, step.name, err)
		}
		if i < len(steps)-1 {
			time.Sleep(2 * time.Second)
		}
	}
	return nil
}

// newScreenSaver constructs a screensaver from application config.
func newScreenSaver(cfg *config.Config, disp display.Display, log *logger.Logger) (*screensaver.ScreenSaver, error) {
	idleTimeout, err := time.ParseDuration(cfg.ScreenSaver.IdleTimeout)
	if err != nil {
		return nil, fmt.Errorf("invalid screensaver.idle_timeout: %w", err)
	}
	ssCfg := screensaver.Config{
		Enabled:          cfg.ScreenSaver.Enabled,
		Mode:             screensaver.Mode(cfg.ScreenSaver.Mode),
		IdleTimeout:      idleTimeout,
		DimBrightness:    cfg.ScreenSaver.DimBrightness,
		NormalBrightness: cfg.ScreenSaver.NormalBrightness,
		ActiveHours: screensaver.ActiveHours{
			Enabled: cfg.ScreenSaver.ActiveHours.Enabled,
			Start:   cfg.ScreenSaver.ActiveHours.Start,
			End:     cfg.ScreenSaver.ActiveHours.End,
		},
	}
	return screensaver.New(ssCfg, disp, log), nil
}
