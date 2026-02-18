package stats

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAvgCollector(t *testing.T) {
	collector := NewLoadAvgCollectorWithPath("../../testdata/proc/loadavg")

	avg1, avg5, avg15, err := collector.GetLoadAvg()
	if err != nil {
		t.Fatalf("GetLoadAvg() failed: %v", err)
	}

	if avg1 < 0.24 || avg1 > 0.26 {
		t.Errorf("expected avg1~0.25, got %f", avg1)
	}
	if avg5 < 0.20 || avg5 > 0.22 {
		t.Errorf("expected avg5~0.21, got %f", avg5)
	}
	if avg15 < 0.26 || avg15 > 0.28 {
		t.Errorf("expected avg15~0.27, got %f", avg15)
	}
}

func TestLoadAvgCollectorNonExistent(t *testing.T) {
	collector := NewLoadAvgCollectorWithPath("/nonexistent/loadavg")

	_, _, _, err := collector.GetLoadAvg()
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestLoadAvgCollectorMalformed(t *testing.T) {
	// Create a temp file with malformed content
	dir := t.TempDir()
	path := filepath.Join(dir, "loadavg")

	tests := []struct {
		name    string
		content string
	}{
		{"too few fields", "0.25 0.21\n"},
		{"non-numeric", "abc def ghi 2/1418 71871\n"},
		{"empty", "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(tt.content), 0o644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			collector := NewLoadAvgCollectorWithPath(path)
			_, _, _, err := collector.GetLoadAvg()
			if err == nil {
				t.Errorf("expected error for malformed input %q", tt.content)
			}
		})
	}
}
