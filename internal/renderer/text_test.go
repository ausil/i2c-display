package renderer

import "testing"

func TestMetricColor(t *testing.T) {
	tests := []struct {
		name    string
		percent float64
		want    string
	}{
		{"low usage", 50, "green"},
		{"zero", 0, "green"},
		{"at 60 boundary", 60, "yellow"},
		{"mid warning", 70, "yellow"},
		{"at 85 boundary", 85, "yellow"},
		{"just above 85", 85.1, "red"},
		{"high usage", 90, "red"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MetricColor(tt.percent)
			var label string
			switch got {
			case ColorGreen:
				label = "green"
			case ColorYellow:
				label = "yellow"
			case ColorRed:
				label = "red"
			default:
				t.Fatalf("unexpected color %v for percent %.1f", got, tt.percent)
			}
			if label != tt.want {
				t.Errorf("MetricColor(%.1f) = %s, want %s", tt.percent, label, tt.want)
			}
		})
	}
}

func TestTempColor(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
		want    string
	}{
		{"cool", 40, "green"},
		{"zero", 0, "green"},
		{"at 55 boundary", 55, "yellow"},
		{"warm", 60, "yellow"},
		{"at 75 boundary", 75, "yellow"},
		{"just above 75", 75.1, "red"},
		{"hot", 80, "red"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TempColor(tt.celsius)
			var label string
			switch got {
			case ColorGreen:
				label = "green"
			case ColorYellow:
				label = "yellow"
			case ColorRed:
				label = "red"
			default:
				t.Fatalf("unexpected color %v for celsius %.1f", got, tt.celsius)
			}
			if label != tt.want {
				t.Errorf("TempColor(%.1f) = %s, want %s", tt.celsius, label, tt.want)
			}
		})
	}
}
