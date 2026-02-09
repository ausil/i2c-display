package config

// DisplaySpec holds the specifications for a display type
type DisplaySpec struct {
	Width  int
	Height int
}

// GetDisplaySpec returns the dimensions for a display type
func GetDisplaySpec(displayType string) (DisplaySpec, bool) {
	specs := map[string]DisplaySpec{
		"ssd1306":        {Width: 128, Height: 64},
		"ssd1306_128x64": {Width: 128, Height: 64},
		"ssd1306_128x32": {Width: 128, Height: 32},
		"ssd1306_96x16":  {Width: 96, Height: 16},
	}

	spec, ok := specs[displayType]
	return spec, ok
}

// ApplyDisplayDefaults applies default width/height based on display type
func (c *DisplayConfig) ApplyDisplayDefaults() {
	if c.Type == "" {
		c.Type = "ssd1306"
	}

	// Get spec for this display type
	spec, ok := GetDisplaySpec(c.Type)
	if !ok {
		return // Unknown type, let validation handle it
	}

	// Auto-fill width and height if not set
	if c.Width == 0 {
		c.Width = spec.Width
	}
	if c.Height == 0 {
		c.Height = spec.Height
	}
}
