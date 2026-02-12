package logger

import (
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	globalLoggerMu sync.RWMutex // Protects global logger operations
)

// Logger wraps zerolog with application-specific configuration
type Logger struct {
	logger zerolog.Logger
}

// Config holds logger configuration
type Config struct {
	Level  string // debug, info, warn, error
	Output string // stdout, stderr
	JSON   bool   // true for JSON output, false for console
}

// New creates a new configured logger
func New(cfg Config) *Logger {
	// Set log level (protected by mutex)
	globalLoggerMu.Lock()
	level := parseLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)
	globalLoggerMu.Unlock()

	// Choose output
	var output io.Writer
	switch strings.ToLower(cfg.Output) {
	case "stderr":
		output = os.Stderr
	default:
		output = os.Stdout
	}

	// Configure format
	var logger zerolog.Logger
	if cfg.JSON {
		logger = zerolog.New(output).With().Timestamp().Logger()
	} else {
		// Pretty console output
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}
		logger = zerolog.New(output).With().Timestamp().Logger()
	}

	return &Logger{logger: logger}
}

// NewDefault creates a logger with default settings
func NewDefault() *Logger {
	return New(Config{
		Level:  "info",
		Output: "stdout",
		JSON:   false,
	})
}

// parseLevel converts string level to zerolog level
func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}

// ErrorWithErr logs an error with the error object
func (l *Logger) ErrorWithErr(err error, msg string) {
	l.logger.Error().Err(err).Msg(msg)
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

// Fatalf logs a formatted fatal error and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal().Msgf(format, args...)
}

// FatalWithErr logs a fatal error with error object and exits
func (l *Logger) FatalWithErr(err error, msg string) {
	l.logger.Fatal().Err(err).Msg(msg)
}

// With adds a field to the logger context
func (l *Logger) With() *Event {
	return &Event{event: l.logger.With()}
}

// Event wraps zerolog context for fluent API
type Event struct {
	event zerolog.Context
}

// Str adds a string field
func (e *Event) Str(key, value string) *Event {
	e.event = e.event.Str(key, value)
	return e
}

// Int adds an int field
func (e *Event) Int(key string, value int) *Event {
	e.event = e.event.Int(key, value)
	return e
}

// Float64 adds a float64 field
func (e *Event) Float64(key string, value float64) *Event {
	e.event = e.event.Float64(key, value)
	return e
}

// Bool adds a boolean field
func (e *Event) Bool(key string, value bool) *Event {
	e.event = e.event.Bool(key, value)
	return e
}

// Err adds an error field
func (e *Event) Err(err error) *Event {
	e.event = e.event.Err(err)
	return e
}

// Logger returns the configured logger
func (e *Event) Logger() *Logger {
	return &Logger{logger: e.event.Logger()}
}

// SetGlobalLogger sets a global logger for use throughout the app
func SetGlobalLogger(l *Logger) {
	globalLoggerMu.Lock()
	defer globalLoggerMu.Unlock()
	log.Logger = l.logger
}

// Global returns the global logger
func Global() *Logger {
	globalLoggerMu.RLock()
	defer globalLoggerMu.RUnlock()
	return &Logger{logger: log.Logger}
}
