package config

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
		"sh1106":         {Width: 128, Height: 64},
		"sh1106_128x64":  {Width: 128, Height: 64},

		// SSD1327 (grayscale) - Driver needed
		"ssd1327":        {Width: 128, Height: 128},
		"ssd1327_128x128": {Width: 128, Height: 128},
		"ssd1327_96x96":  {Width: 96, Height: 96},

		// SSD1331 (color OLED) - Driver needed
		"ssd1331":        {Width: 96, Height: 64},
		"ssd1331_96x64":  {Width: 96, Height: 64},
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
}
