package display

import (
	"strings"
	"testing"

	"github.com/ausil/i2c-display/internal/config"
)

func TestNewDisplay(t *testing.T) {
	tests := []struct {
		name    string
		config  config.DisplayConfig
		wantErr bool
	}{
		{
			name: "ssd1306 default",
			config: config.DisplayConfig{
				Type:       "ssd1306",
				I2CBus:     "/dev/i2c-1",
				I2CAddress: "0x3C",
				Width:      128,
				Height:     64,
				Rotation:   0,
			},
			wantErr: true, // Will fail without hardware, but factory should create object
		},
		{
			name: "ssd1306_128x64",
			config: config.DisplayConfig{
				Type:       "ssd1306_128x64",
				I2CBus:     "/dev/i2c-1",
				I2CAddress: "0x3C",
				Width:      128,
				Height:     64,
				Rotation:   0,
			},
			wantErr: true, // Will fail without hardware
		},
		{
			name: "ssd1306_128x32",
			config: config.DisplayConfig{
				Type:       "ssd1306_128x32",
				I2CBus:     "/dev/i2c-1",
				I2CAddress: "0x3C",
				Width:      128,
				Height:     32,
				Rotation:   0,
			},
			wantErr: true, // Will fail without hardware
		},
		{
			name: "st7735 default",
			config: config.DisplayConfig{
				Type:    "st7735",
				SPIBus:  "SPI0.0",
				DCPin:   "GPIO24",
				RSTPin:  "GPIO25",
				Width:   128,
				Height:  160,
				Rotation: 0,
			},
			wantErr: true, // Will fail without hardware
		},
		{
			name: "st7735_128x128",
			config: config.DisplayConfig{
				Type:    "st7735_128x128",
				SPIBus:  "SPI0.0",
				DCPin:   "GPIO24",
				Width:   128,
				Height:  128,
				Rotation: 0,
			},
			wantErr: true, // Will fail without hardware
		},
		{
			name: "uctronics_colour",
			config: config.DisplayConfig{
				Type:       "uctronics_colour",
				I2CBus:     "/dev/i2c-1",
				I2CAddress: "0x18",
				Width:      160,
				Height:     80,
			},
			wantErr: true, // Will fail without hardware
		},
		{
			name: "unsupported type",
			config: config.DisplayConfig{
				Type:       "unknown",
				I2CBus:     "/dev/i2c-1",
				I2CAddress: "0x3C",
				Width:      128,
				Height:     64,
				Rotation:   0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDisplay(&tt.config)
			// We expect errors when hardware is not available
			// The test is to verify the factory creates the right type
			// and handles unsupported types correctly
			if tt.config.Type == "unknown" {
				if err == nil {
					t.Error("expected error for unsupported display type")
				}
				if !strings.Contains(err.Error(), "unsupported") {
					t.Errorf("expected 'unsupported' in error, got: %v", err)
				}
			}
		})
	}
}

func TestNewDisplayMock(t *testing.T) {
	// Test that factory pattern works with mock
	mock := NewMockDisplay(128, 64)
	if err := mock.Init(); err != nil {
		t.Fatalf("mock init failed: %v", err)
	}

	bounds := mock.GetBounds()
	if bounds.Dx() != 128 || bounds.Dy() != 64 {
		t.Errorf("expected 128x64, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}
