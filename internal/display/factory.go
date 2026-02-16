package display

import (
	"fmt"
	"strings"

	"github.com/ausil/i2c-display/internal/config"
)

// NewDisplay creates a display implementation based on configuration
func NewDisplay(cfg *config.DisplayConfig) (Display, error) {
	displayType := strings.ToLower(cfg.Type)

	// SSD1306 variants (official periph.io support)
	if strings.HasPrefix(displayType, "ssd1306") {
		return NewSSD1306Display(
			cfg.I2CBus,
			cfg.I2CAddress,
			cfg.Width,
			cfg.Height,
			cfg.Rotation,
		)
	}

	// ST7735 variants (SPI TFT)
	if strings.HasPrefix(displayType, "st7735") {
		return NewST7735Display(
			cfg.SPIBus,
			cfg.DCPin,
			cfg.RSTPin,
			cfg.Width,
			cfg.Height,
			cfg.Rotation,
			displayType,
		)
	}

	// Other display types - Framework ready, awaiting drivers
	supportedButNeedDrivers := map[string]string{
		"sh1106":  "SH1106 (128x64 mono) - compatible with SSD1306, driver available at github.com/danielgatis/go-sh1106 (SPI)",
		"ssd1327": "SSD1327 (128x128 grayscale) - no Go I2C driver found",
		"ssd1331": "SSD1331 (96x64 color) - no Go I2C driver found",
	}

	for prefix, desc := range supportedButNeedDrivers {
		if strings.HasPrefix(displayType, prefix) {
			return nil, fmt.Errorf("display type %s is recognized but not yet implemented: %s\n"+
				"See DISPLAY_TYPES.md for how to add this display", displayType, desc)
		}
	}

	return nil, fmt.Errorf("unsupported display type: %s", cfg.Type)
}
