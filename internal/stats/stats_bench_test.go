package stats

import (
	"testing"

	"github.com/ausil/i2c-display/internal/config"
)

func BenchmarkSystemCollector(b *testing.B) {
	cfg := config.Default()
	collector, err := NewSystemCollector(cfg)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := collector.Collect(); err != nil {
			b.Fatal(err)
		}
	}
}
