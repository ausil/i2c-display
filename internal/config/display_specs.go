package config

import "strings"

// DisplaySpec holds the specifications for a display type
type DisplaySpec struct {
	Width  int
	Height int
}

// GetDisplaySpec returns the dimensions for a display type
func GetDisplaySpec(displayType string) (DisplaySpec, bool) {
	specs := map[string]DisplaySpec{
		// SSD1306 family (fully supported via periph.io)
		"ssd1306":        {Width: 128, Height: 64},
		"ssd1306_128x64": {Width: 128, Height: 64},
		"ssd1306_128x32": {Width: 128, Height: 32},
		"ssd1306_96x16":  {Width: 96, Height: 16},

		// SH1106 family (via third-party driver)
		"sh1106":        {Width: 128, Height: 64},
		"sh1106_128x64": {Width: 128, Height: 64},

		// SSD1327 (grayscale) - Driver needed
		"ssd1327":         {Width: 128, Height: 128},
		"ssd1327_128x128": {Width: 128, Height: 128},
		"ssd1327_96x96":   {Width: 96, Height: 96},

		// SSD1331 (color OLED) - Driver needed
		"ssd1331":       {Width: 96, Height: 64},
		"ssd1331_96x64": {Width: 96, Height: 64},

		// ST7735 (color TFT via SPI)
		"st7735":         {Width: 128, Height: 160},
		"st7735_128x160": {Width: 128, Height: 160},
		"st7735_128x128": {Width: 128, Height: 128},
		"st7735_160x80":  {Width: 160, Height: 80},

		// UCTRONICS (I2C-bridged ST7735 via onboard MCU)
		"uctronics_colour": {Width: 160, Height: 80},
	}

	spec, ok := specs[displayType]
	return spec, ok
}

// ApplyDisplayDefaults applies default width/height based on display type
// The display type is authoritative - dimensions are always set to match the type
func (c *DisplayConfig) ApplyDisplayDefaults() {
	if c.Type == "" {
		c.Type = "ssd1306"
	}

	// Get spec for this display type
	spec, ok := GetDisplaySpec(c.Type)
	if !ok {
		return // Unknown type, let validation handle it
	}

	// Always set width and height to match the display type
	// This makes the type authoritative over explicit dimension values
	c.Width = spec.Width
	c.Height = spec.Height

	// UCTRONICS displays use an I2C bridge MCU at address 0x18
	if strings.HasPrefix(strings.ToLower(c.Type), "uctronics") {
		c.I2CAddress = "0x18"
		if c.I2CBus == "" {
			c.I2CBus = "/dev/i2c-1"
		}
	}
}
