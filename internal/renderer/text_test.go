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

func TestLoadColor(t *testing.T) {
	tests := []struct {
		name   string
		load   float64
		numCPU int
		want   string
	}{
		{"low load 4 cores", 1.0, 4, "green"},           // 0.25 per core
		{"zero load", 0, 4, "green"},                     // 0.0 per core
		{"yellow boundary 4 cores", 2.8, 4, "yellow"},    // 0.7 per core
		{"mid warning 4 cores", 3.0, 4, "yellow"},        // 0.75 per core
		{"at red boundary 4 cores", 4.0, 4, "yellow"},    // 1.0 per core
		{"above red 4 cores", 4.1, 4, "red"},             // 1.025 per core
		{"high load 4 cores", 5.0, 4, "red"},             // 1.25 per core
		{"single core green", 0.5, 1, "green"},           // 0.5 per core
		{"single core yellow", 0.7, 1, "yellow"},         // 0.7 per core
		{"single core red", 1.1, 1, "red"},               // 1.1 per core
		{"zero numCPU defaults to 1", 0.5, 0, "green"},   // treated as 1 core
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LoadColor(tt.load, tt.numCPU)
			var label string
			switch got {
			case ColorGreen:
				label = "green"
			case ColorYellow:
				label = "yellow"
			case ColorRed:
				label = "red"
			default:
				t.Fatalf("unexpected color %v for load %.1f, numCPU %d", got, tt.load, tt.numCPU)
			}
			if label != tt.want {
				t.Errorf("LoadColor(%.1f, %d) = %s, want %s", tt.load, tt.numCPU, label, tt.want)
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
