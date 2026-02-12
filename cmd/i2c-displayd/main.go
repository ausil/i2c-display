package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/ausil/i2c-display/internal/config"
	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/logger"
	"github.com/ausil/i2c-display/internal/renderer"
	"github.com/ausil/i2c-display/internal/rotation"
	"github.com/ausil/i2c-display/internal/stats"
)

//nolint:funlen // main function naturally has many statements for initialization
func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to configuration file")
	useMock := flag.Bool("mock", false, "Use mock display (for testing without hardware)")
	validateConfig := flag.Bool("validate-config", false, "Validate configuration and exit")
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

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
			cfg = newCfg
			log.Info("Configuration reloaded successfully")
			continue

		case syscall.SIGINT, syscall.SIGTERM:
			log.With().Str("signal", sig.String()).Logger().Info("Received shutdown signal")
			goto shutdown
		}
	}

shutdown:

	// Cancel context to stop rotation manager
	cancel()

	// Stop manager gracefully
	mgr.Stop()

	log.Info("Shutdown complete")
}

