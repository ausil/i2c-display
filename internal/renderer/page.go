package renderer

import (
	"github.com/denniskorablev/ssd1306-display/internal/display"
	"github.com/denniskorablev/ssd1306-display/internal/stats"
)

// Page represents a displayable page
type Page interface {
	// Render draws the page to the display
	Render(disp display.Display, stats *stats.SystemStats) error

	// Title returns a short title for the page
	Title() string
}
