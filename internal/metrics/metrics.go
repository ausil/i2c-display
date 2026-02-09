package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ausil/i2c-display/internal/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector holds all Prometheus metrics for the application
type Collector struct {
	// Display metrics
	DisplayRefreshTotal   *prometheus.CounterVec
	DisplayRefreshErrors  *prometheus.CounterVec
	DisplayRefreshLatency *prometheus.HistogramVec

	// I2C metrics
	I2CErrorsTotal *prometheus.CounterVec

	// System metrics
	CPUTemperature     prometheus.Gauge
	MemoryUsedPercent  prometheus.Gauge
	DiskUsedPercent    prometheus.Gauge
	NetworkInterfaces  prometheus.Gauge

	// Page metrics
	CurrentPage        prometheus.Gauge
	PageRotationTotal  prometheus.Counter

	registry *prometheus.Registry
	log      *logger.Logger
}

// Config holds metrics server configuration
type Config struct {
	Enabled bool   `json:"enabled"`
	Address string `json:"address"` // e.g., ":9090"
}

// New creates a new metrics collector
func New(log *logger.Logger) *Collector {
	registry := prometheus.NewRegistry()

	c := &Collector{
		DisplayRefreshTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "i2c_display_refresh_total",
				Help: "Total number of display refreshes",
			},
			[]string{"status"}, // success or error
		),
		DisplayRefreshErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "i2c_display_refresh_errors_total",
				Help: "Total number of display refresh errors by type",
			},
			[]string{"error_type"},
		),
		DisplayRefreshLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "i2c_display_refresh_latency_seconds",
				Help:    "Histogram of display refresh latencies in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"page_type"}, // system or network
		),
		I2CErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "i2c_display_i2c_errors_total",
				Help: "Total number of I2C communication errors",
			},
			[]string{"operation"}, // init, show, etc.
		),
		CPUTemperature: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "i2c_display_cpu_temperature_celsius",
				Help: "CPU temperature in Celsius",
			},
		),
		MemoryUsedPercent: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "i2c_display_memory_used_percent",
				Help: "Memory usage percentage",
			},
		),
		DiskUsedPercent: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "i2c_display_disk_used_percent",
				Help: "Disk usage percentage",
			},
		),
		NetworkInterfaces: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "i2c_display_network_interfaces_count",
				Help: "Number of detected network interfaces",
			},
		),
		CurrentPage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "i2c_display_current_page",
				Help: "Current display page number",
			},
		),
		PageRotationTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "i2c_display_page_rotation_total",
				Help: "Total number of page rotations",
			},
		),
		registry: registry,
		log:      log,
	}

	// Register all metrics
	registry.MustRegister(
		c.DisplayRefreshTotal,
		c.DisplayRefreshErrors,
		c.DisplayRefreshLatency,
		c.I2CErrorsTotal,
		c.CPUTemperature,
		c.MemoryUsedPercent,
		c.DiskUsedPercent,
		c.NetworkInterfaces,
		c.CurrentPage,
		c.PageRotationTotal,
	)

	return c
}

// RecordDisplayRefresh records a display refresh operation
func (c *Collector) RecordDisplayRefresh(success bool, duration time.Duration, pageType string) {
	if success {
		c.DisplayRefreshTotal.WithLabelValues("success").Inc()
	} else {
		c.DisplayRefreshTotal.WithLabelValues("error").Inc()
	}
	c.DisplayRefreshLatency.WithLabelValues(pageType).Observe(duration.Seconds())
}

// RecordDisplayError records a display error
func (c *Collector) RecordDisplayError(errorType string) {
	c.DisplayRefreshErrors.WithLabelValues(errorType).Inc()
}

// RecordI2CError records an I2C communication error
func (c *Collector) RecordI2CError(operation string) {
	c.I2CErrorsTotal.WithLabelValues(operation).Inc()
}

// UpdateSystemMetrics updates system stat metrics
func (c *Collector) UpdateSystemMetrics(cpuTemp float64, memPercent float64, diskPercent float64, interfaceCount int) {
	if cpuTemp > 0 {
		c.CPUTemperature.Set(cpuTemp)
	}
	c.MemoryUsedPercent.Set(memPercent)
	c.DiskUsedPercent.Set(diskPercent)
	c.NetworkInterfaces.Set(float64(interfaceCount))
}

// RecordPageRotation records a page rotation
func (c *Collector) RecordPageRotation(pageNum int) {
	c.PageRotationTotal.Inc()
	c.CurrentPage.Set(float64(pageNum))
}

// Server wraps the HTTP server for metrics
type Server struct {
	httpServer *http.Server
	log        *logger.Logger
}

// NewServer creates a new metrics HTTP server
func NewServer(cfg Config, collector *Collector, log *logger.Logger) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(collector.registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Address,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		log: log,
	}
}

// Start starts the metrics server
func (s *Server) Start() error {
	s.log.With().Str("address", s.httpServer.Addr).Logger().Info("Starting metrics server")

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.ErrorWithErr(err, "Metrics server error")
		}
	}()

	return nil
}

// Stop gracefully stops the metrics server
func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("Stopping metrics server")
	return s.httpServer.Shutdown(ctx)
}

// StartMetricsServer starts the Prometheus metrics server if enabled
func StartMetricsServer(cfg Config, collector *Collector, log *logger.Logger) (*Server, error) {
	if !cfg.Enabled {
		log.Debug("Metrics server disabled")
		return nil, nil
	}

	server := NewServer(cfg, collector, log)
	if err := server.Start(); err != nil {
		return nil, fmt.Errorf("failed to start metrics server: %w", err)
	}

	return server, nil
}
