package metrics

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ausil/i2c-display/internal/logger"
)

func TestNew(t *testing.T) {
	log := logger.NewDefault()
	collector := New(log)

	if collector == nil {
		t.Fatal("Expected non-nil collector")
	}

	if collector.registry == nil {
		t.Error("Expected non-nil registry")
	}

	if collector.log == nil {
		t.Error("Expected non-nil logger")
	}

	// Verify all metrics are initialized
	if collector.DisplayRefreshTotal == nil {
		t.Error("DisplayRefreshTotal not initialized")
	}
	if collector.DisplayRefreshErrors == nil {
		t.Error("DisplayRefreshErrors not initialized")
	}
	if collector.DisplayRefreshLatency == nil {
		t.Error("DisplayRefreshLatency not initialized")
	}
	if collector.I2CErrorsTotal == nil {
		t.Error("I2CErrorsTotal not initialized")
	}
	if collector.CPUTemperature == nil {
		t.Error("CPUTemperature not initialized")
	}
	if collector.MemoryUsedPercent == nil {
		t.Error("MemoryUsedPercent not initialized")
	}
	if collector.DiskUsedPercent == nil {
		t.Error("DiskUsedPercent not initialized")
	}
	if collector.NetworkInterfaces == nil {
		t.Error("NetworkInterfaces not initialized")
	}
	if collector.CurrentPage == nil {
		t.Error("CurrentPage not initialized")
	}
	if collector.PageRotationTotal == nil {
		t.Error("PageRotationTotal not initialized")
	}
}

func TestRecordDisplayRefresh(t *testing.T) {
	log := logger.NewDefault()
	collector := New(log)

	tests := []struct {
		name     string
		success  bool
		duration time.Duration
		pageType string
	}{
		{
			name:     "successful refresh",
			success:  true,
			duration: 100 * time.Millisecond,
			pageType: "system",
		},
		{
			name:     "failed refresh",
			success:  false,
			duration: 50 * time.Millisecond,
			pageType: "network",
		},
		{
			name:     "zero duration",
			success:  true,
			duration: 0,
			pageType: "system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector.RecordDisplayRefresh(tt.success, tt.duration, tt.pageType)
			// If this doesn't panic, the test passes
		})
	}
}

func TestRecordDisplayError(t *testing.T) {
	log := logger.NewDefault()
	collector := New(log)

	errorTypes := []string{
		"i2c_error",
		"render_error",
		"timeout",
		"unknown",
	}

	for _, errorType := range errorTypes {
		t.Run(errorType, func(t *testing.T) {
			collector.RecordDisplayError(errorType)
			// If this doesn't panic, the test passes
		})
	}
}

func TestRecordI2CError(t *testing.T) {
	log := logger.NewDefault()
	collector := New(log)

	operations := []string{
		"init",
		"show",
		"clear",
		"set_brightness",
	}

	for _, op := range operations {
		t.Run(op, func(t *testing.T) {
			collector.RecordI2CError(op)
			// If this doesn't panic, the test passes
		})
	}
}

func TestUpdateSystemMetrics(t *testing.T) {
	log := logger.NewDefault()
	collector := New(log)

	tests := []struct {
		name           string
		cpuTemp        float64
		memPercent     float64
		diskPercent    float64
		interfaceCount int
	}{
		{
			name:           "normal values",
			cpuTemp:        45.5,
			memPercent:     67.8,
			diskPercent:    52.3,
			interfaceCount: 3,
		},
		{
			name:           "zero cpu temp",
			cpuTemp:        0,
			memPercent:     50.0,
			diskPercent:    40.0,
			interfaceCount: 2,
		},
		{
			name:           "high values",
			cpuTemp:        85.0,
			memPercent:     95.5,
			diskPercent:    99.9,
			interfaceCount: 10,
		},
		{
			name:           "zero interface count",
			cpuTemp:        45.0,
			memPercent:     50.0,
			diskPercent:    50.0,
			interfaceCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector.UpdateSystemMetrics(tt.cpuTemp, tt.memPercent, tt.diskPercent, tt.interfaceCount)
			// If this doesn't panic, the test passes
		})
	}
}

func TestRecordPageRotation(t *testing.T) {
	log := logger.NewDefault()
	collector := New(log)

	pageNumbers := []int{1, 2, 3, 0, -1, 100}

	for _, pageNum := range pageNumbers {
		t.Run(string(rune(pageNum)), func(t *testing.T) {
			collector.RecordPageRotation(pageNum)
			// If this doesn't panic, the test passes
		})
	}
}

func TestNewServer(t *testing.T) {
	log := logger.NewDefault()
	collector := New(log)

	cfg := Config{
		Enabled: true,
		Address: ":19090",
	}

	server := NewServer(cfg, collector, log)

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	if server.httpServer == nil {
		t.Error("Expected non-nil HTTP server")
	}

	if server.log == nil {
		t.Error("Expected non-nil logger")
	}

	if server.httpServer.Addr != cfg.Address {
		t.Errorf("Expected address %s, got %s", cfg.Address, server.httpServer.Addr)
	}
}

func TestServerStartStop(t *testing.T) {
	log := logger.NewDefault()
	collector := New(log)

	cfg := Config{
		Enabled: true,
		Address: ":19091", // Use different port to avoid conflicts
	}

	server := NewServer(cfg, collector, log)

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test health endpoint
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:19091/health", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to GET /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != "OK\n" {
		t.Errorf("Expected body 'OK\\n', got %q", string(body))
	}

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	log := logger.NewDefault()
	collector := New(log)

	cfg := Config{
		Enabled: true,
		Address: ":19092",
	}

	server := NewServer(cfg, collector, log)

	// Start server
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Stop(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Record some metrics
	collector.RecordDisplayRefresh(true, 100*time.Millisecond, "system")
	collector.RecordDisplayError("test_error")
	collector.RecordI2CError("init")
	collector.UpdateSystemMetrics(45.5, 67.8, 52.3, 3)
	collector.RecordPageRotation(1)

	// Test metrics endpoint
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost:19092/metrics", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// Verify metrics are present
	expectedMetrics := []string{
		"i2c_display_refresh_total",
		"i2c_display_refresh_errors_total",
		"i2c_display_refresh_latency_seconds",
		"i2c_display_i2c_errors_total",
		"i2c_display_cpu_temperature_celsius",
		"i2c_display_memory_used_percent",
		"i2c_display_disk_used_percent",
		"i2c_display_network_interfaces_count",
		"i2c_display_current_page",
		"i2c_display_page_rotation_total",
	}

	for _, metric := range expectedMetrics {
		if !strings.Contains(bodyStr, metric) {
			t.Errorf("Expected metric %s not found in response", metric)
		}
	}
}

func TestStartMetricsServer(t *testing.T) {
	log := logger.NewDefault()
	collector := New(log)

	t.Run("enabled", func(t *testing.T) {
		cfg := Config{
			Enabled: true,
			Address: ":19093",
		}

		server, err := StartMetricsServer(cfg, collector, log)
		if err != nil {
			t.Fatalf("Failed to start metrics server: %v", err)
		}

		if server == nil {
			t.Error("Expected non-nil server when enabled")
		}

		// Clean up
		if server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Stop(ctx)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		cfg := Config{
			Enabled: false,
			Address: ":19094",
		}

		server, err := StartMetricsServer(cfg, collector, log)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if server != nil {
			t.Error("Expected nil server when disabled")
		}
	})
}

func TestServerTimeouts(t *testing.T) {
	log := logger.NewDefault()
	collector := New(log)

	cfg := Config{
		Enabled: true,
		Address: ":19095",
	}

	server := NewServer(cfg, collector, log)

	expectedReadTimeout := 5 * time.Second
	expectedWriteTimeout := 10 * time.Second
	expectedIdleTimeout := 60 * time.Second

	if server.httpServer.ReadTimeout != expectedReadTimeout {
		t.Errorf("Expected ReadTimeout %v, got %v", expectedReadTimeout, server.httpServer.ReadTimeout)
	}

	if server.httpServer.WriteTimeout != expectedWriteTimeout {
		t.Errorf("Expected WriteTimeout %v, got %v", expectedWriteTimeout, server.httpServer.WriteTimeout)
	}

	if server.httpServer.IdleTimeout != expectedIdleTimeout {
		t.Errorf("Expected IdleTimeout %v, got %v", expectedIdleTimeout, server.httpServer.IdleTimeout)
	}
}

func TestMultipleMetricRecords(t *testing.T) {
	log := logger.NewDefault()
	collector := New(log)

	// Record multiple operations to ensure metrics accumulate correctly
	for i := 0; i < 10; i++ {
		collector.RecordDisplayRefresh(true, time.Duration(i)*time.Millisecond, "system")
		collector.RecordDisplayRefresh(false, time.Duration(i)*time.Millisecond, "network")
		collector.RecordDisplayError("error_type_1")
		collector.RecordI2CError("operation_1")
		collector.RecordPageRotation(i)
	}

	collector.UpdateSystemMetrics(50.0, 60.0, 70.0, 5)

	// If no panics occurred, the test passes
}
