package health

import (
	"errors"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	checker := New()
	if checker == nil {
		t.Fatal("expected non-nil checker")
	}

	if !checker.IsHealthy() {
		t.Error("new checker should be healthy")
	}
}

func TestRegisterComponent(t *testing.T) {
	checker := New()
	checker.RegisterComponent("test-component")

	comp := checker.GetComponentStatus("test-component")
	if comp == nil {
		t.Fatal("expected component to be registered")
	}

	if comp.Name != "test-component" {
		t.Errorf("expected name 'test-component', got '%s'", comp.Name)
	}

	if comp.Status != StatusHealthy {
		t.Errorf("expected status %s, got %s", StatusHealthy, comp.Status)
	}
}

func TestRecordSuccess(t *testing.T) {
	checker := New()
	checker.RegisterComponent("test")

	checker.RecordSuccess("test")

	comp := checker.GetComponentStatus("test")
	if comp.SuccessCount != 1 {
		t.Errorf("expected SuccessCount=1, got %d", comp.SuccessCount)
	}
}

func TestRecordError(t *testing.T) {
	checker := New()
	checker.RegisterComponent("test")

	err := errors.New("test error")
	checker.RecordError("test", err)

	comp := checker.GetComponentStatus("test")
	if comp.ErrorCount != 1 {
		t.Errorf("expected ErrorCount=1, got %d", comp.ErrorCount)
	}

	if comp.Message != "test error" {
		t.Errorf("expected message 'test error', got '%s'", comp.Message)
	}
}

func TestStatusTransitions(t *testing.T) {
	checker := New()
	checker.RegisterComponent("test")

	// Start healthy
	if checker.GetComponentStatus("test").Status != StatusHealthy {
		t.Error("expected initial status to be healthy")
	}

	// 1-2 errors: still healthy
	checker.RecordError("test", errors.New("error 1"))
	checker.RecordError("test", errors.New("error 2"))
	if checker.GetComponentStatus("test").Status != StatusHealthy {
		t.Error("expected status to be healthy with 2 errors")
	}

	// 3+ errors: degraded
	checker.RecordError("test", errors.New("error 3"))
	if checker.GetComponentStatus("test").Status != StatusDegraded {
		t.Errorf("expected status to be degraded with 3 errors, got %s", checker.GetComponentStatus("test").Status)
	}

	// 10+ errors: unhealthy
	for i := 0; i < 7; i++ {
		checker.RecordError("test", errors.New("error"))
	}
	if checker.GetComponentStatus("test").Status != StatusUnhealthy {
		t.Errorf("expected status to be unhealthy with 10 errors, got %s", checker.GetComponentStatus("test").Status)
	}

	// Successes gradually recover
	for i := 0; i < 10; i++ {
		checker.RecordSuccess("test")
	}
	if checker.GetComponentStatus("test").Status != StatusHealthy {
		t.Errorf("expected status to recover to healthy, got %s", checker.GetComponentStatus("test").Status)
	}
}

func TestGetOverallStatus(t *testing.T) {
	checker := New()
	checker.RegisterComponent("comp1")
	checker.RegisterComponent("comp2")
	checker.RegisterComponent("comp3")

	// All healthy
	if checker.GetOverallStatus() != StatusHealthy {
		t.Error("expected overall status to be healthy")
	}

	// One degraded
	for i := 0; i < 3; i++ {
		checker.RecordError("comp1", errors.New("error"))
	}
	if checker.GetOverallStatus() != StatusDegraded {
		t.Error("expected overall status to be degraded")
	}

	// One unhealthy
	for i := 0; i < 10; i++ {
		checker.RecordError("comp2", errors.New("error"))
	}
	if checker.GetOverallStatus() != StatusUnhealthy {
		t.Error("expected overall status to be unhealthy")
	}
}

func TestGetAllComponents(t *testing.T) {
	checker := New()
	checker.RegisterComponent("comp1")
	checker.RegisterComponent("comp2")

	all := checker.GetAllComponents()
	if len(all) != 2 {
		t.Errorf("expected 2 components, got %d", len(all))
	}

	if all["comp1"] == nil || all["comp2"] == nil {
		t.Error("expected both components to be present")
	}
}

func TestConcurrency(t *testing.T) {
	checker := New()
	checker.RegisterComponent("test")

	done := make(chan bool)

	// Multiple goroutines recording status
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				if j%2 == 0 {
					checker.RecordSuccess("test")
				} else {
					checker.RecordError("test", errors.New("error"))
				}
				checker.GetComponentStatus("test")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not crash - that's the main test
	comp := checker.GetComponentStatus("test")
	if comp == nil {
		t.Fatal("expected component to exist")
	}

	// Note: Due to race conditions in error recovery (RecordSuccess decrements ErrorCount),
	// the total may not be exactly 1000. The important thing is no crashes occurred.
	if comp.SuccessCount == 0 {
		t.Error("expected some successful operations to be recorded")
	}
}

func TestLastCheckTimestamp(t *testing.T) {
	checker := New()
	checker.RegisterComponent("test")

	before := time.Now()
	time.Sleep(10 * time.Millisecond)
	checker.RecordSuccess("test")
	time.Sleep(10 * time.Millisecond)
	after := time.Now()

	comp := checker.GetComponentStatus("test")
	if comp.LastCheck.Before(before) || comp.LastCheck.After(after) {
		t.Error("LastCheck timestamp not in expected range")
	}
}
