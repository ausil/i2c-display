package logger

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "info level stdout",
			config: Config{
				Level:  "info",
				Output: "stdout",
				JSON:   false,
			},
		},
		{
			name: "debug level stderr json",
			config: Config{
				Level:  "debug",
				Output: "stderr",
				JSON:   true,
			},
		},
		{
			name: "warn level",
			config: Config{
				Level:  "warn",
				Output: "stdout",
				JSON:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.config)
			if logger == nil {
				t.Fatal("expected logger, got nil")
			}
		})
	}
}

func TestNewDefault(t *testing.T) {
	logger := NewDefault()
	if logger == nil {
		t.Fatal("expected logger, got nil")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected zerolog.Level
	}{
		{"debug", zerolog.DebugLevel},
		{"info", zerolog.InfoLevel},
		{"warn", zerolog.WarnLevel},
		{"warning", zerolog.WarnLevel},
		{"error", zerolog.ErrorLevel},
		{"invalid", zerolog.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLevel(tt.input)
			if result != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		logger: zerolog.New(&buf),
	}

	// Test each log method
	logger.Debug("debug message")
	logger.Debugf("debug %s", "formatted")
	logger.Info("info message")
	logger.Infof("info %s", "formatted")
	logger.Warn("warn message")
	logger.Warnf("warn %s", "formatted")
	logger.Error("error message")
	logger.Errorf("error %s", "formatted")

	// Check that something was logged
	if buf.Len() == 0 {
		t.Error("expected log output, got none")
	}
}

func TestLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	baseLogger := &Logger{
		logger: zerolog.New(&buf),
	}

	contextLogger := baseLogger.With().
		Str("component", "test").
		Int("count", 42).
		Logger()

	contextLogger.Info("test message")

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("component")) {
		t.Error("expected context field 'component' in output")
	}
}
