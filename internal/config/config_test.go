package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Display.I2CBus != "/dev/i2c-1" {
		t.Errorf("expected I2CBus=/dev/i2c-1, got %s", cfg.Display.I2CBus)
	}
	if cfg.Display.Width != 128 {
		t.Errorf("expected Width=128, got %d", cfg.Display.Width)
	}
	if cfg.Display.Height != 64 {
		t.Errorf("expected Height=64, got %d", cfg.Display.Height)
	}
	if cfg.Pages.RotationInterval != "5s" {
		t.Errorf("expected RotationInterval=5s, got %s", cfg.Pages.RotationInterval)
	}
	if cfg.Network.MaxInterfacesPerPage != 3 {
		t.Errorf("expected MaxInterfacesPerPage=3, got %d", cfg.Network.MaxInterfacesPerPage)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid default config",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name: "empty i2c bus",
			modify: func(c *Config) {
				c.Display.I2CBus = ""
			},
			wantErr: true,
			errMsg:  "i2c_bus cannot be empty",
		},
		{
			name: "empty i2c address",
			modify: func(c *Config) {
				c.Display.I2CAddress = ""
			},
			wantErr: true,
			errMsg:  "i2c_address cannot be empty",
		},
		{
			name: "invalid width",
			modify: func(c *Config) {
				c.Display.Width = 0
			},
			wantErr: true,
			errMsg:  "width must be positive",
		},
		{
			name: "invalid height",
			modify: func(c *Config) {
				c.Display.Height = -1
			},
			wantErr: true,
			errMsg:  "height must be positive",
		},
		{
			name: "invalid rotation",
			modify: func(c *Config) {
				c.Display.Rotation = 4
			},
			wantErr: true,
			errMsg:  "rotation must be 0-3",
		},
		{
			name: "invalid rotation interval",
			modify: func(c *Config) {
				c.Pages.RotationInterval = "invalid"
			},
			wantErr: true,
			errMsg:  "invalid pages.rotation_interval",
		},
		{
			name: "invalid refresh interval",
			modify: func(c *Config) {
				c.Pages.RefreshInterval = "not-a-duration"
			},
			wantErr: true,
			errMsg:  "invalid pages.refresh_interval",
		},
		{
			name: "invalid hostname display",
			modify: func(c *Config) {
				c.SystemInfo.HostnameDisplay = "invalid"
			},
			wantErr: true,
			errMsg:  "hostname_display must be 'short' or 'full'",
		},
		{
			name: "empty disk path",
			modify: func(c *Config) {
				c.SystemInfo.DiskPath = ""
			},
			wantErr: true,
			errMsg:  "disk_path cannot be empty",
		},
		{
			name: "invalid temperature unit",
			modify: func(c *Config) {
				c.SystemInfo.TemperatureUnit = "kelvin"
			},
			wantErr: true,
			errMsg:  "temperature_unit must be 'celsius' or 'fahrenheit'",
		},
		{
			name: "invalid max interfaces per page",
			modify: func(c *Config) {
				c.Network.MaxInterfacesPerPage = 0
			},
			wantErr: true,
			errMsg:  "max_interfaces_per_page must be positive",
		},
		{
			name: "invalid log level",
			modify: func(c *Config) {
				c.Logging.Level = "verbose"
			},
			wantErr: true,
			errMsg:  "logging.level must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Default()
			tt.modify(cfg)

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, should contain %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestGetIntervals(t *testing.T) {
	cfg := Default()

	rotInterval, err := cfg.Pages.GetRotationInterval()
	if err != nil {
		t.Fatalf("GetRotationInterval() error = %v", err)
	}
	if rotInterval.Seconds() != 5 {
		t.Errorf("expected 5 seconds, got %v", rotInterval)
	}

	refInterval, err := cfg.Pages.GetRefreshInterval()
	if err != nil {
		t.Fatalf("GetRefreshInterval() error = %v", err)
	}
	if refInterval.Seconds() != 1 {
		t.Errorf("expected 1 second, got %v", refInterval)
	}
}

func TestLoad(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	validConfig := `{
		"display": {
			"i2c_bus": "/dev/i2c-1",
			"i2c_address": "0x3C",
			"width": 128,
			"height": 64,
			"rotation": 0
		},
		"pages": {
			"rotation_interval": "10s",
			"refresh_interval": "2s"
		},
		"system_info": {
			"hostname_display": "short",
			"disk_path": "/",
			"temperature_source": "/sys/class/thermal/thermal_zone0/temp",
			"temperature_unit": "celsius"
		},
		"network": {
			"auto_detect": true,
			"interface_filter": {
				"include": ["eth0"],
				"exclude": ["lo"]
			},
			"show_ipv4": true,
			"show_ipv6": false,
			"max_interfaces_per_page": 3
		},
		"logging": {
			"level": "debug",
			"output": "stdout"
		}
	}`

	invalidJSON := `{"display": invalid json`

	invalidConfig := `{
		"display": {
			"i2c_bus": "",
			"i2c_address": "0x3C",
			"width": 128,
			"height": 64,
			"rotation": 0
		},
		"pages": {
			"rotation_interval": "5s",
			"refresh_interval": "1s"
		},
		"system_info": {
			"hostname_display": "short",
			"disk_path": "/",
			"temperature_source": "/sys/class/thermal/thermal_zone0/temp",
			"temperature_unit": "celsius"
		},
		"network": {
			"auto_detect": true,
			"interface_filter": {
				"include": [],
				"exclude": []
			},
			"show_ipv4": true,
			"show_ipv6": false,
			"max_interfaces_per_page": 3
		},
		"logging": {
			"level": "info",
			"output": "stdout"
		}
	}`

	configWithSsd1306_128x32 := `{
		"display": {
			"type": "ssd1306_128x32",
			"i2c_bus": "/dev/i2c-1",
			"i2c_address": "0x3C",
			"rotation": 0
		},
		"pages": {
			"rotation_interval": "5s",
			"refresh_interval": "1s"
		},
		"system_info": {
			"hostname_display": "short",
			"disk_path": "/",
			"temperature_source": "/sys/class/thermal/thermal_zone0/temp",
			"temperature_unit": "celsius"
		},
		"network": {
			"auto_detect": true,
			"interface_filter": {
				"include": ["eth0"],
				"exclude": ["lo"]
			},
			"show_ipv4": true,
			"show_ipv6": false,
			"max_interfaces_per_page": 3
		},
		"logging": {
			"level": "info",
			"output": "stdout"
		}
	}`

	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(*testing.T, *Config)
	}{
		{
			name:    "valid config",
			content: validConfig,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Pages.RotationInterval != "10s" {
					t.Errorf("expected RotationInterval=10s, got %s", cfg.Pages.RotationInterval)
				}
				if cfg.Logging.Level != "debug" {
					t.Errorf("expected Level=debug, got %s", cfg.Logging.Level)
				}
			},
		},
		{
			name:    "ssd1306_128x32 without dimensions",
			content: configWithSsd1306_128x32,
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.Display.Type != "ssd1306_128x32" {
					t.Errorf("expected Type=ssd1306_128x32, got %s", cfg.Display.Type)
				}
				if cfg.Display.Width != 128 {
					t.Errorf("expected Width=128 (auto-filled from type), got %d", cfg.Display.Width)
				}
				if cfg.Display.Height != 32 {
					t.Errorf("expected Height=32 (auto-filled from type), got %d", cfg.Display.Height)
				}
			},
		},
		{
			name:    "invalid json",
			content: invalidJSON,
			wantErr: true,
		},
		{
			name:    "invalid config",
			content: invalidConfig,
			wantErr: true,
		},
		{
			name:    "non-existent file",
			content: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.content != "" {
				path = filepath.Join(tmpDir, tt.name+".json")
				if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			} else {
				path = filepath.Join(tmpDir, "nonexistent.json")
			}

			cfg, err := Load(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func TestLoadWithPriority(t *testing.T) {
	tmpDir := t.TempDir()

	validConfig := `{
		"display": {
			"i2c_bus": "/dev/i2c-1",
			"i2c_address": "0x3C",
			"width": 128,
			"height": 64,
			"rotation": 0
		},
		"pages": {
			"rotation_interval": "5s",
			"refresh_interval": "1s"
		},
		"system_info": {
			"hostname_display": "short",
			"disk_path": "/",
			"temperature_source": "/sys/class/thermal/thermal_zone0/temp",
			"temperature_unit": "celsius"
		},
		"network": {
			"auto_detect": true,
			"interface_filter": {
				"include": ["eth0"],
				"exclude": ["lo"]
			},
			"show_ipv4": true,
			"show_ipv6": false,
			"max_interfaces_per_page": 3
		},
		"logging": {
			"level": "info",
			"output": "stdout"
		}
	}`

	// Test with explicit path
	explicitPath := filepath.Join(tmpDir, "explicit.json")
	if err := os.WriteFile(explicitPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("failed to write explicit config: %v", err)
	}

	cfg, err := LoadWithPriority(explicitPath)
	if err != nil {
		t.Fatalf("LoadWithPriority() with explicit path error = %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}

	// Test with environment variable
	envPath := filepath.Join(tmpDir, "env.json")
	if err := os.WriteFile(envPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("failed to write env config: %v", err)
	}
	os.Setenv("I2C_DISPLAY_CONFIG_PATH", envPath)
	defer os.Unsetenv("I2C_DISPLAY_CONFIG_PATH")

	cfg, err = LoadWithPriority("")
	if err != nil {
		t.Fatalf("LoadWithPriority() with env var error = %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}

	// Test with no valid paths
	os.Unsetenv("I2C_DISPLAY_CONFIG_PATH")
	_, err = LoadWithPriority("")
	if err == nil {
		t.Error("LoadWithPriority() should fail when no config found")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
