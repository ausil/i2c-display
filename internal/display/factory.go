package display

import (
	"fmt"
	"strings"

	"github.com/ausil/i2c-display/internal/config"
)

// NewDisplay creates a display implementation based on configuration
func NewDisplay(cfg *config.DisplayConfig) (Display, error) {
	displayType := strings.ToLower(cfg.Type)

	// All SSD1306 variants use the same driver
	if strings.HasPrefix(displayType, "ssd1306") {
		return NewSSD1306Display(
			cfg.I2CBus,
			cfg.I2CAddress,
			cfg.Width,
			cfg.Height,
			cfg.Rotation,
		)
	}

	return nil, fmt.Errorf("unsupported display type: %s (currently only SSD1306 variants are supported)", cfg.Type)
}
