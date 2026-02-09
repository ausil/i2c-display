package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoSuccess(t *testing.T) {
	cfg := DefaultConfig()
	ctx := context.Background()

	callCount := 0
	err := Do(ctx, cfg, func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestDoRetryAndSuccess(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InitialDelay = 10 * time.Millisecond
	cfg.MaxDelay = 50 * time.Millisecond
	ctx := context.Background()

	callCount := 0
	err := Do(ctx, cfg, func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestDoMaxAttempts(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InitialDelay = 10 * time.Millisecond
	cfg.MaxAttempts = 3
	ctx := context.Background()

	callCount := 0
	err := Do(ctx, cfg, func() error {
		callCount++
		return errors.New("persistent error")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}

	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestDoContextCancellation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InitialDelay = 100 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after first failure
	callCount := 0
	err := Do(ctx, cfg, func() error {
		callCount++
		if callCount == 1 {
			cancel()
		}
		return errors.New("error")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}

	if callCount > 2 {
		t.Errorf("expected at most 2 calls, got %d", callCount)
	}
}

func TestDoWithResultSuccess(t *testing.T) {
	cfg := DefaultConfig()
	ctx := context.Background()

	result, err := DoWithResult(ctx, cfg, func() (int, error) {
		return 42, nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestDoWithResultRetry(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InitialDelay = 10 * time.Millisecond
	ctx := context.Background()

	callCount := 0
	result, err := DoWithResult(ctx, cfg, func() (string, error) {
		callCount++
		if callCount < 2 {
			return "", errors.New("temporary error")
		}
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("expected 'success', got '%s'", result)
	}
}

func TestExponentialBackoff(t *testing.T) {
	cfg := Config{
		MaxAttempts:  5,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}
	ctx := context.Background()

	startTime := time.Now()
	callTimes := []time.Time{}

	Do(ctx, cfg, func() error {
		callTimes = append(callTimes, time.Now())
		return errors.New("always fail")
	})

	elapsed := time.Since(startTime)

	// Should have at least some delay (conservative check)
	if elapsed < 100*time.Millisecond {
		t.Errorf("expected at least 100ms total delay, got %v", elapsed)
	}

	// Check that delays increase
	if len(callTimes) >= 3 {
		delay1 := callTimes[1].Sub(callTimes[0])
		delay2 := callTimes[2].Sub(callTimes[1])
		if delay2 <= delay1 {
			t.Errorf("expected increasing delays, got %v and %v", delay1, delay2)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", cfg.MaxAttempts)
	}

	if cfg.InitialDelay != 100*time.Millisecond {
		t.Errorf("expected InitialDelay=100ms, got %v", cfg.InitialDelay)
	}

	if cfg.MaxDelay != 5*time.Second {
		t.Errorf("expected MaxDelay=5s, got %v", cfg.MaxDelay)
	}

	if cfg.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %f", cfg.Multiplier)
	}
}
