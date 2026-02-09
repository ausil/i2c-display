package retry

import (
	"context"
	"fmt"
	"time"
)

// Config holds retry configuration
type Config struct {
	MaxAttempts     int           // Maximum number of retry attempts
	InitialDelay    time.Duration // Initial delay between retries
	MaxDelay        time.Duration // Maximum delay between retries
	Multiplier      float64       // Backoff multiplier (typically 2.0)
	RetryableErrors []string      // List of error substrings that trigger retry
}

// DefaultConfig returns a sensible default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// Operation is a function that can be retried
type Operation func() error

// Do executes an operation with exponential backoff retry logic
func Do(ctx context.Context, cfg Config, op Operation) error {
	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Try the operation
		if err := op(); err != nil {
			lastErr = err

			// Check if we should retry
			if attempt >= cfg.MaxAttempts {
				return fmt.Errorf("operation failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
			}

			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return fmt.Errorf("operation cancelled: %w", ctx.Err())
			default:
			}

			// Wait before retrying with exponential backoff
			select {
			case <-ctx.Done():
				return fmt.Errorf("operation cancelled during backoff: %w", ctx.Err())
			case <-time.After(delay):
			}

			// Calculate next delay with exponential backoff
			delay = time.Duration(float64(delay) * cfg.Multiplier)
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}

			continue
		}

		// Operation succeeded
		return nil
	}

	return lastErr
}

// DoWithResult executes an operation with exponential backoff and returns a result
func DoWithResult[T any](ctx context.Context, cfg Config, op func() (T, error)) (T, error) {
	var result T
	var lastErr error
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Try the operation
		res, err := op()
		if err != nil {
			lastErr = err

			// Check if we should retry
			if attempt >= cfg.MaxAttempts {
				return result, fmt.Errorf("operation failed after %d attempts: %w", cfg.MaxAttempts, lastErr)
			}

			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return result, fmt.Errorf("operation cancelled: %w", ctx.Err())
			default:
			}

			// Wait before retrying with exponential backoff
			select {
			case <-ctx.Done():
				return result, fmt.Errorf("operation cancelled during backoff: %w", ctx.Err())
			case <-time.After(delay):
			}

			// Calculate next delay with exponential backoff
			delay = time.Duration(float64(delay) * cfg.Multiplier)
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}

			continue
		}

		// Operation succeeded
		return res, nil
	}

	return result, lastErr
}
