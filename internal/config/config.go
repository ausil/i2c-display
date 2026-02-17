package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config represents the application configuration
type Config struct {
	Display     DisplayConfig     `json:"display"`
	Pages       PagesConfig       `json:"pages"`
	SystemInfo  SystemInfoConfig  `json:"system_info"`
	Network     NetworkConfig     `json:"network"`
	Logging     LoggingConfig     `json:"logging"`
	Metrics     MetricsConfig     `json:"metrics"`
	ScreenSaver ScreenSaverConfig `json:"screensaver"`
}

// DisplayConfig holds display-related settings
type DisplayConfig struct {
	Type       string `json:"type"`        // Display type: "ssd1306", "ssd1306_128x32", "st7735", etc.
	I2CBus     string `json:"i2c_bus"`
	I2CAddress string `json:"i2c_address"`
	SPIBus     string `json:"spi_bus"`
	DCPin      string `json:"dc_pin"`
	RSTPin     string `json:"rst_pin"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Rotation   int    `json:"rotation"`
}

// IsI2C returns true if this display connects via I2C
func (c *DisplayConfig) IsI2C() bool {
	t := strings.ToLower(c.Type)
	return strings.HasPrefix(t, "ssd1306") ||
		strings.HasPrefix(t, "sh1106") ||
		strings.HasPrefix(t, "ssd1327") ||
		strings.HasPrefix(t, "ssd1331") ||
		strings.HasPrefix(t, "uctronics")
}

// IsSPI returns true if this display connects via SPI
func (c *DisplayConfig) IsSPI() bool {
	return strings.HasPrefix(strings.ToLower(c.Type), "st7735")
}

// PagesConfig holds page rotation settings
type PagesConfig struct {
	RotationInterval string `json:"rotation_interval"`
	RefreshInterval  string `json:"refresh_interval"`
}

// SystemInfoConfig holds system information settings
type SystemInfoConfig struct {
	HostnameDisplay   string `json:"hostname_display"`
	DiskPath          string `json:"disk_path"`
	TemperatureSource string `json:"temperature_source"`
	TemperatureUnit   string `json:"temperature_unit"`
}

// NetworkConfig holds network interface settings
type NetworkConfig struct {
	AutoDetect            bool            `json:"auto_detect"`
	InterfaceFilter       InterfaceFilter `json:"interface_filter"`
	ShowIPv4              bool            `json:"show_ipv4"`
	ShowIPv6              bool            `json:"show_ipv6"`
	MaxInterfacesPerPage  int             `json:"max_interfaces_per_page"`
}

// InterfaceFilter defines include/exclude patterns for network interfaces
type InterfaceFilter struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level  string `json:"level"`
	Output string `json:"output"`
	JSON   bool   `json:"json"` // true for JSON output, false for console
}

// MetricsConfig holds Prometheus metrics settings
type MetricsConfig struct {
	Enabled bool   `json:"enabled"`
	Address string `json:"address"` // e.g., "127.0.0.1:9090"
}

// ScreenSaverConfig holds screen saver settings
type ScreenSaverConfig struct {
	Enabled          bool   `json:"enabled"`
	Mode             string `json:"mode"`              // "off", "dim", or "blank"
	IdleTimeout      string `json:"idle_timeout"`      // e.g., "5m"
	DimBrightness    uint8  `json:"dim_brightness"`    // 0-255
	NormalBrightness uint8  `json:"normal_brightness"` // 0-255
}

// GetRotationInterval returns the parsed rotation interval duration
func (p *PagesConfig) GetRotationInterval() (time.Duration, error) {
	return time.ParseDuration(p.RotationInterval)
}

// GetRefreshInterval returns the parsed refresh interval duration
func (p *PagesConfig) GetRefreshInterval() (time.Duration, error) {
	return time.ParseDuration(p.RefreshInterval)
}

// Default returns a configuration with sensible defaults
func Default() *Config {
	cfg := &Config{
		Display: DisplayConfig{
			Type:       "ssd1306",
			I2CBus:     "/dev/i2c-1",
			I2CAddress: "0x3C",
			Width:      0, // Will be set by ApplyDisplayDefaults based on type
			Height:     0, // Will be set by ApplyDisplayDefaults based on type
			Rotation:   0,
		},
		Pages: PagesConfig{
			RotationInterval: "5s",
			RefreshInterval:  "1s",
		},
		SystemInfo: SystemInfoConfig{
			HostnameDisplay:   "short",
			DiskPath:          "/",
			TemperatureSource: "/sys/class/thermal/thermal_zone0/temp",
			TemperatureUnit:   "celsius",
		},
		Network: NetworkConfig{
			AutoDetect: true,
			InterfaceFilter: InterfaceFilter{
				Include: []string{"eth0", "wlan0", "usb0"},
				Exclude: []string{"lo", "docker*", "veth*"},
			},
			ShowIPv4:             true,
			ShowIPv6:             false,
			MaxInterfacesPerPage: 3,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Output: "stdout",
			JSON:   false,
		},
		Metrics: MetricsConfig{
			Enabled: false,
			Address: "127.0.0.1:9090",
		},
		ScreenSaver: ScreenSaverConfig{
			Enabled:          false,
			Mode:             "dim",
			IdleTimeout:      "5m",
			DimBrightness:    50,
			NormalBrightness: 255,
		},
	}

	// Apply display defaults based on type
	cfg.Display.ApplyDisplayDefaults()

	return cfg
}

// Load loads configuration from a file path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply display defaults based on type
	cfg.Display.ApplyDisplayDefaults()

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// LoadWithPriority loads configuration using cascading priority:
// 1. Explicit path (if provided and exists)
// 2. I2C_DISPLAY_CONFIG_PATH environment variable
// 3. /etc/i2c-display/config.json
// 4. $HOME/.config/i2c-display/config.json
// 5. ./config.json
func LoadWithPriority(explicitPath string) (*Config, error) {
	var paths []string

	// Priority 1: Explicit path
	if explicitPath != "" {
		paths = append(paths, explicitPath)
	}

	// Priority 2: Environment variable
	if envPath := os.Getenv("I2C_DISPLAY_CONFIG_PATH"); envPath != "" {
		paths = append(paths, envPath)
	}

	// Priority 3: System-wide
	paths = append(paths, "/etc/i2c-display/config.json")

	// Priority 4: User-specific
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "i2c-display", "config.json"))
	}

	// Priority 5: Current directory
	paths = append(paths, "./config.json")

	var lastErr error
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			cfg, err := Load(path)
			if err != nil {
				lastErr = fmt.Errorf("%s: %w", path, err)
				continue
			}
			return cfg, nil
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fmt.Errorf("no configuration file found in search paths: %v", paths)
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	if err := c.validateDisplay(); err != nil {
		return err
	}
	if err := c.validatePages(); err != nil {
		return err
	}
	if err := c.validateSystemInfo(); err != nil {
		return err
	}
	if err := c.validateNetwork(); err != nil {
		return err
	}
	if err := c.validateLogging(); err != nil {
		return err
	}
	if err := c.validateScreenSaver(); err != nil {
		return err
	}
	return c.validateMetrics()
}

func (c *Config) validateDisplay() error {
	if c.Display.Type == "" {
		c.Display.Type = "ssd1306" // Default to SSD1306
	}

	spec, validType := GetDisplaySpec(c.Display.Type)
	if !validType {
		return fmt.Errorf("display.type %q is not a recognized display type", c.Display.Type)
	}

	if c.Display.IsI2C() {
		if c.Display.I2CBus == "" {
			return fmt.Errorf("display.i2c_bus cannot be empty")
		}
		if !strings.HasPrefix(c.Display.I2CBus, "/") {
			return fmt.Errorf("display.i2c_bus must be an absolute path, got %s", c.Display.I2CBus)
		}
		if c.Display.I2CAddress == "" {
			return fmt.Errorf("display.i2c_address cannot be empty")
		}
		addrLower := strings.ToLower(c.Display.I2CAddress)
		if !strings.HasPrefix(addrLower, "0x") {
			return fmt.Errorf("display.i2c_address must be in hex format (e.g., 0x3C), got %s", c.Display.I2CAddress)
		}
		if _, err := strconv.ParseUint(addrLower[2:], 16, 8); err != nil {
			return fmt.Errorf("display.i2c_address is not a valid 8-bit hex address (e.g., 0x3C), got %s", c.Display.I2CAddress)
		}
	}

	if c.Display.IsSPI() {
		if c.Display.SPIBus == "" {
			return fmt.Errorf("display.spi_bus cannot be empty for SPI display type %s", c.Display.Type)
		}
		if c.Display.DCPin == "" {
			return fmt.Errorf("display.dc_pin cannot be empty for SPI display type %s", c.Display.Type)
		}
	}

	if c.Display.Width <= 0 {
		return fmt.Errorf("display.width must be positive, got %d", c.Display.Width)
	}
	if c.Display.Height <= 0 {
		return fmt.Errorf("display.height must be positive, got %d", c.Display.Height)
	}

	if c.Display.Width != spec.Width || c.Display.Height != spec.Height {
		return fmt.Errorf("display dimensions (%dx%d) don't match type %s (expected %dx%d)",
			c.Display.Width, c.Display.Height, c.Display.Type, spec.Width, spec.Height)
	}

	if c.Display.Rotation < 0 || c.Display.Rotation > 3 {
		return fmt.Errorf("display.rotation must be 0-3, got %d", c.Display.Rotation)
	}

	return nil
}

func (c *Config) validatePages() error {
	if _, err := c.Pages.GetRotationInterval(); err != nil {
		return fmt.Errorf("invalid pages.rotation_interval: %w", err)
	}
	if _, err := c.Pages.GetRefreshInterval(); err != nil {
		return fmt.Errorf("invalid pages.refresh_interval: %w", err)
	}
	return nil
}

func (c *Config) validateSystemInfo() error {
	if c.SystemInfo.HostnameDisplay != "short" && c.SystemInfo.HostnameDisplay != "full" {
		return fmt.Errorf("system_info.hostname_display must be 'short' or 'full', got %s", c.SystemInfo.HostnameDisplay)
	}
	if c.SystemInfo.DiskPath == "" {
		return fmt.Errorf("system_info.disk_path cannot be empty")
	}
	if _, err := os.Stat(c.SystemInfo.DiskPath); err != nil {
		return fmt.Errorf("system_info.disk_path %q does not exist: %w", c.SystemInfo.DiskPath, err)
	}
	if c.SystemInfo.TemperatureUnit != "celsius" && c.SystemInfo.TemperatureUnit != "fahrenheit" {
		return fmt.Errorf("system_info.temperature_unit must be 'celsius' or 'fahrenheit', got %s", c.SystemInfo.TemperatureUnit)
	}
	return nil
}

func (c *Config) validateNetwork() error {
	if c.Network.MaxInterfacesPerPage <= 0 {
		return fmt.Errorf("network.max_interfaces_per_page must be positive, got %d", c.Network.MaxInterfacesPerPage)
	}
	for _, pattern := range c.Network.InterfaceFilter.Include {
		if _, err := filepath.Match(pattern, ""); err != nil {
			return fmt.Errorf("network.interface_filter.include contains invalid glob pattern %q: %w", pattern, err)
		}
	}
	for _, pattern := range c.Network.InterfaceFilter.Exclude {
		if _, err := filepath.Match(pattern, ""); err != nil {
			return fmt.Errorf("network.interface_filter.exclude contains invalid glob pattern %q: %w", pattern, err)
		}
	}
	return nil
}

func (c *Config) validateLogging() error {
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("logging.level must be one of [debug, info, warn, error], got %s", c.Logging.Level)
	}
	return nil
}

func (c *Config) validateScreenSaver() error {
	if !c.ScreenSaver.Enabled {
		return nil
	}

	validModes := map[string]bool{"off": true, "dim": true, "blank": true}
	if !validModes[c.ScreenSaver.Mode] {
		return fmt.Errorf("screensaver.mode must be one of [off, dim, blank], got %s", c.ScreenSaver.Mode)
	}

	d, err := time.ParseDuration(c.ScreenSaver.IdleTimeout)
	if err != nil {
		return fmt.Errorf("screensaver.idle_timeout is not a valid duration: %w", err)
	}
	if d <= 0 {
		return fmt.Errorf("screensaver.idle_timeout must be positive, got %s", c.ScreenSaver.IdleTimeout)
	}

	if c.ScreenSaver.Mode == "dim" && c.ScreenSaver.DimBrightness >= c.ScreenSaver.NormalBrightness {
		return fmt.Errorf("screensaver.dim_brightness (%d) must be less than normal_brightness (%d)",
			c.ScreenSaver.DimBrightness, c.ScreenSaver.NormalBrightness)
	}

	return nil
}

func (c *Config) validateMetrics() error {
	if !c.Metrics.Enabled {
		return nil
	}

	if c.Metrics.Address == "" {
		return fmt.Errorf("metrics.address cannot be empty when metrics are enabled")
	}

	return nil
}
