package config

import "testing"

func TestGetDisplaySpec(t *testing.T) {
	tests := []struct {
		name        string
		displayType string
		wantWidth   int
		wantHeight  int
		wantOK      bool
	}{
		{
			name:        "ssd1306 default",
			displayType: "ssd1306",
			wantWidth:   128,
			wantHeight:  64,
			wantOK:      true,
		},
		{
			name:        "ssd1306_128x64",
			displayType: "ssd1306_128x64",
			wantWidth:   128,
			wantHeight:  64,
			wantOK:      true,
		},
		{
			name:        "ssd1306_128x32",
			displayType: "ssd1306_128x32",
			wantWidth:   128,
			wantHeight:  32,
			wantOK:      true,
		},
		{
			name:        "ssd1306_96x16",
			displayType: "ssd1306_96x16",
			wantWidth:   96,
			wantHeight:  16,
			wantOK:      true,
		},
		{
			name:        "st7735 default",
			displayType: "st7735",
			wantWidth:   128,
			wantHeight:  160,
			wantOK:      true,
		},
		{
			name:        "st7735_128x160",
			displayType: "st7735_128x160",
			wantWidth:   128,
			wantHeight:  160,
			wantOK:      true,
		},
		{
			name:        "st7735_128x128",
			displayType: "st7735_128x128",
			wantWidth:   128,
			wantHeight:  128,
			wantOK:      true,
		},
		{
			name:        "st7735_160x80",
			displayType: "st7735_160x80",
			wantWidth:   160,
			wantHeight:  80,
			wantOK:      true,
		},
		{
			name:        "uctronics_colour",
			displayType: "uctronics_colour",
			wantWidth:   160,
			wantHeight:  80,
			wantOK:      true,
		},
		{
			name:        "unknown type",
			displayType: "unknown",
			wantWidth:   0,
			wantHeight:  0,
			wantOK:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, ok := GetDisplaySpec(tt.displayType)
			if ok != tt.wantOK {
				t.Errorf("GetDisplaySpec() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok {
				if spec.Width != tt.wantWidth {
					t.Errorf("GetDisplaySpec() width = %v, want %v", spec.Width, tt.wantWidth)
				}
				if spec.Height != tt.wantHeight {
					t.Errorf("GetDisplaySpec() height = %v, want %v", spec.Height, tt.wantHeight)
				}
			}
		})
	}
}

func TestApplyDisplayDefaults(t *testing.T) {
	tests := []struct {
		name   string
		config DisplayConfig
		want   DisplayConfig
	}{
		{
			name: "auto-fill dimensions for ssd1306",
			config: DisplayConfig{
				Type:   "ssd1306",
				Width:  0,
				Height: 0,
			},
			want: DisplayConfig{
				Type:   "ssd1306",
				Width:  128,
				Height: 64,
			},
		},
		{
			name: "auto-fill dimensions for ssd1306_128x32",
			config: DisplayConfig{
				Type:   "ssd1306_128x32",
				Width:  0,
				Height: 0,
			},
			want: DisplayConfig{
				Type:   "ssd1306_128x32",
				Width:  128,
				Height: 32,
			},
		},
		{
			name: "override dimensions to match type",
			config: DisplayConfig{
				Type:   "ssd1306",
				Width:  96, // Wrong dimensions
				Height: 16, // Will be overridden
			},
			want: DisplayConfig{
				Type:   "ssd1306",
				Width:  128, // Corrected to match type
				Height: 64,  // Corrected to match type
			},
		},
		{
			name: "correct wrong dimensions for type",
			config: DisplayConfig{
				Type:   "ssd1306_128x32",
				Width:  128,
				Height: 64, // Wrong for this type
			},
			want: DisplayConfig{
				Type:   "ssd1306_128x32",
				Width:  128,
				Height: 32, // Corrected to match type
			},
		},
		{
			name: "default type to ssd1306",
			config: DisplayConfig{
				Type:   "",
				Width:  0,
				Height: 0,
			},
			want: DisplayConfig{
				Type:   "ssd1306",
				Width:  128,
				Height: 64,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.ApplyDisplayDefaults()

			if tt.config.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", tt.config.Type, tt.want.Type)
			}
			if tt.config.Width != tt.want.Width {
				t.Errorf("Width = %v, want %v", tt.config.Width, tt.want.Width)
			}
			if tt.config.Height != tt.want.Height {
				t.Errorf("Height = %v, want %v", tt.config.Height, tt.want.Height)
			}
		})
	}
}
