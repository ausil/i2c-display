package renderer

import (
	"fmt"
	"image"
	"image/color"

	"github.com/ausil/i2c-display/internal/display"
	"github.com/ausil/i2c-display/internal/stats"
)

const loadHistorySize = 300 // 5 minutes at 1s refresh

// LoadGraphPage displays a rolling graph of system load average
type LoadGraphPage struct {
	history []float64 // ring buffer, max loadHistorySize entries
	head    int       // next write position
	count   int       // number of valid entries
	numCPU  int       // cached CPU count for scaling
	lines   int       // configured line count (0=auto, 2=default, 4=compact)
}

// NewLoadGraphPage creates a new load graph page
func NewLoadGraphPage(lines int) *LoadGraphPage {
	return &LoadGraphPage{
		history: make([]float64, loadHistorySize),
		lines:   lines,
	}
}

// Title returns the page title
func (p *LoadGraphPage) Title() string {
	return "Load"
}

// Render draws the load graph page
func (p *LoadGraphPage) Render(disp display.Display, s *stats.SystemStats) error {
	// Record current load into ring buffer
	p.history[p.head] = s.LoadAvg1
	p.head = (p.head + 1) % loadHistorySize
	if p.count < loadHistorySize {
		p.count++
	}

	// Cache CPU count
	if s.NumCPU > 0 {
		p.numCPU = s.NumCPU
	}
	if p.numCPU <= 0 {
		p.numCPU = 1
	}

	if err := disp.Clear(); err != nil {
		return err
	}

	bounds := disp.GetBounds()
	layout := NewLayout(bounds, p.lines)

	// Header
	if layout.ShowHeader {
		if err := DrawTextCenteredColorScaled(disp, layout.HeaderY, s.Hostname, ColorGreen, layout.TextScale); err != nil {
			return err
		}
	}

	// Separator
	if layout.ShowSeparator {
		if err := DrawLine(disp, layout.SeparatorY); err != nil {
			return err
		}
	}

	// Small displays: text only
	if layout.Height <= 32 {
		return p.renderSmall(disp, s, layout)
	}

	return p.renderGraph(disp, s, layout, bounds)
}

// renderSmall renders text-only output for small displays (height <= 32)
func (p *LoadGraphPage) renderSmall(disp display.Display, s *stats.SystemStats, layout *Layout) error {
	if len(layout.ContentLines) == 0 {
		return disp.Show()
	}

	text := fmt.Sprintf("L:%.2f %.2f %.2f", s.LoadAvg1, s.LoadAvg5, s.LoadAvg15)
	maxWidth := layout.Width - 2*MarginLeft
	if layout.TextScale > 0 && layout.TextScale < 1 {
		text = TruncateTextSmall(text, maxWidth)
	} else {
		text = TruncateText(text, maxWidth)
	}
	c := LoadColor(s.LoadAvg1, p.numCPU)

	if err := DrawTextColorScaled(disp, MarginLeft, layout.ContentLines[0], text, c, layout.TextScale); err != nil {
		return err
	}

	return disp.Show()
}

// renderGraph renders the full graph view for standard displays
func (p *LoadGraphPage) renderGraph(disp display.Display, s *stats.SystemStats, layout *Layout, bounds image.Rectangle) error {
	if len(layout.ContentLines) == 0 {
		return disp.Show()
	}

	// Text label on first content line
	label := fmt.Sprintf("1m:%.2f 5m:%.2f 15m:%.2f", s.LoadAvg1, s.LoadAvg5, s.LoadAvg15)
	maxWidth := bounds.Dx() - 2*MarginLeft
	label = TruncateText(label, maxWidth)
	c := LoadColor(s.LoadAvg1, p.numCPU)

	if err := DrawTextColorScaled(disp, MarginLeft, layout.ContentLines[0], label, c, layout.TextScale); err != nil {
		return err
	}

	// Graph area: below the text label (account for scaled text height)
	textHeight := ScaledTextHeight(layout.TextScale)
	graphY := layout.ContentLines[0] + textHeight + 2
	graphX := MarginLeft
	graphWidth := bounds.Dx() - 2*MarginLeft
	graphHeight := bounds.Dy() - graphY - 1

	if graphWidth <= 0 || graphHeight <= 0 {
		return disp.Show()
	}

	// Build the graph image
	graphImg := p.buildGraphImage(graphWidth, graphHeight)

	if err := disp.DrawImage(graphX, graphY, graphImg); err != nil {
		return err
	}

	return disp.Show()
}

// buildGraphImage creates an NRGBA image of the load graph
func (p *LoadGraphPage) buildGraphImage(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	numCPU := float64(p.numCPU)

	// Determine Y-axis max: at least numCPU, or the max observed load
	yMax := numCPU
	samples := p.getSamples()
	for _, v := range samples {
		if v > yMax {
			yMax = v
		}
	}
	// Add 10% headroom
	yMax *= 1.1
	if yMax <= 0 {
		yMax = 1.0
	}

	// Draw threshold lines (dotted)
	yellowThresh := 0.7 * numCPU
	redThresh := 1.0 * numCPU

	yellowY := height - 1 - int(yellowThresh/yMax*float64(height-1))
	redY := height - 1 - int(redThresh/yMax*float64(height-1))

	for x := 0; x < width; x++ {
		if x%4 < 2 { // dotted pattern
			if yellowY >= 0 && yellowY < height {
				img.SetNRGBA(x, yellowY, color.NRGBA{R: 128, G: 128, B: 0, A: 128})
			}
			if redY >= 0 && redY < height {
				img.SetNRGBA(x, redY, color.NRGBA{R: 128, G: 0, B: 0, A: 128})
			}
		}
	}

	if len(samples) == 0 {
		return img
	}

	// Downsample or right-align
	var bars []float64
	if len(samples) > width {
		// Average groups of samples
		bars = make([]float64, width)
		samplesPerBar := float64(len(samples)) / float64(width)
		for col := 0; col < width; col++ {
			startF := float64(col) * samplesPerBar
			endF := float64(col+1) * samplesPerBar
			start := int(startF)
			end := int(endF)
			if end > len(samples) {
				end = len(samples)
			}
			if start >= end {
				if start < len(samples) {
					bars[col] = samples[start]
				}
				continue
			}
			sum := 0.0
			for i := start; i < end; i++ {
				sum += samples[i]
			}
			bars[col] = sum / float64(end-start)
		}
	} else {
		// Right-align: newest samples at the right edge
		bars = make([]float64, width)
		offset := width - len(samples)
		for i, v := range samples {
			bars[offset+i] = v
		}
	}

	// Draw bars
	for col := 0; col < width; col++ {
		val := bars[col]
		if val <= 0 {
			continue
		}

		barHeight := int(val / yMax * float64(height-1))
		if barHeight <= 0 {
			barHeight = 1
		}
		if barHeight > height {
			barHeight = height
		}

		for row := 0; row < barHeight; row++ {
			y := height - 1 - row
			// Color based on the Y value (what load level this pixel represents)
			pixelLoad := float64(row) / float64(height-1) * yMax
			var clr color.NRGBA
			perCore := pixelLoad / numCPU
			switch {
			case perCore > 1.0:
				clr = ColorRed
			case perCore >= 0.7:
				clr = ColorYellow
			default:
				clr = ColorGreen
			}
			img.SetNRGBA(col, y, clr)
		}
	}

	return img
}

// getSamples returns the history samples in chronological order
func (p *LoadGraphPage) getSamples() []float64 {
	if p.count == 0 {
		return nil
	}

	samples := make([]float64, p.count)
	if p.count < loadHistorySize {
		// Buffer hasn't wrapped yet: samples are 0..count-1
		copy(samples, p.history[:p.count])
	} else {
		// Buffer has wrapped: oldest is at head, newest at head-1
		n := copy(samples, p.history[p.head:])
		copy(samples[n:], p.history[:p.head])
	}
	return samples
}
