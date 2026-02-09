package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ausil/i2c-display/internal/config"
	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/renderer"
	"github.com/ausil/i2c-display/internal/rotation"
	"github.com/ausil/i2c-display/internal/stats"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to configuration file")
	useMock := flag.Bool("mock", false, "Use mock display (for testing without hardware)")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadWithPriority(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set up logging
	setupLogging(cfg)

	log.Println("SSD1306 Display Service starting...")
	log.Printf("Hostname display mode: %s", cfg.SystemInfo.HostnameDisplay)

	// Create display
	var disp display.Display
	if *useMock {
		log.Println("Using mock display (no hardware)")
		disp = display.NewMockDisplay(cfg.Display.Width, cfg.Display.Height)
	} else {
		log.Printf("Initializing SSD1306 display on %s at %s", cfg.Display.I2CBus, cfg.Display.I2CAddress)
		hardwareDisp, err := display.NewSSD1306Display(
			cfg.Display.I2CBus,
			cfg.Display.I2CAddress,
			cfg.Display.Width,
			cfg.Display.Height,
			cfg.Display.Rotation,
		)
		if err != nil {
			log.Printf("Warning: Failed to initialize hardware display: %v", err)
			log.Println("Falling back to mock display")
			disp = display.NewMockDisplay(cfg.Display.Width, cfg.Display.Height)
		} else {
			disp = hardwareDisp
		}
	}

	// Initialize display
	if err := disp.Init(); err != nil {
		log.Fatalf("Failed to initialize display: %v", err)
	}
	defer func() {
		log.Println("Closing display...")
		if err := disp.Close(); err != nil {
			log.Printf("Error closing display: %v", err)
		}
	}()

	// Create stats collector
	collector, err := stats.NewSystemCollector(cfg)
	if err != nil {
		log.Fatalf("Failed to create stats collector: %v", err)
	}

	// Create renderer
	rend := renderer.NewRenderer(disp, cfg)

	// Collect initial stats to build pages
	initialStats, err := collector.Collect()
	if err != nil {
		log.Fatalf("Failed to collect initial stats: %v", err)
	}
	rend.BuildPages(initialStats)

	log.Printf("Built %d page(s)", rend.PageCount())

	// Create rotation manager
	mgr := rotation.NewManager(cfg, collector, rend)

	// Set up context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start rotation manager
	if err := mgr.Start(ctx); err != nil {
		log.Fatalf("Failed to start rotation manager: %v", err)
	}

	log.Println("Display service running. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Received shutdown signal...")

	// Cancel context to stop rotation manager
	cancel()

	// Stop manager gracefully
	mgr.Stop()

	log.Println("Shutdown complete")
}

// setupLogging configures logging based on config
func setupLogging(cfg *config.Config) {
	// For now, just use standard log output
	// In a full implementation, you could configure different log levels
	switch cfg.Logging.Level {
	case "debug":
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	case "info", "warn", "error":
		log.SetFlags(log.Ldate | log.Ltime)
	default:
		log.SetFlags(log.Ldate | log.Ltime)
	}

	if cfg.Logging.Output == "stdout" {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(os.Stderr)
	}
}
