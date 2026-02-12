package health

import (
	"sync"
	"time"
)

// Status represents the health status
type Status string

const (
	// StatusHealthy indicates system is working normally
	StatusHealthy Status = "healthy"
	// StatusDegraded indicates system is working with issues
	StatusDegraded Status = "degraded"
	// StatusUnhealthy indicates system is not functioning
	StatusUnhealthy Status = "unhealthy"
)

// Component represents a system component with health status
type Component struct {
	Name         string    `json:"name"`
	Status       Status    `json:"status"`
	Message      string    `json:"message,omitempty"`
	LastCheck    time.Time `json:"last_check"`
	ErrorCount   int       `json:"error_count"`
	SuccessCount int       `json:"success_count"`
}

// Checker tracks health status of system components
type Checker struct {
	mu         sync.RWMutex
	components map[string]*Component
}

// New creates a new health checker
func New() *Checker {
	return &Checker{
		components: make(map[string]*Component),
	}
}

// RegisterComponent registers a component for health tracking
func (h *Checker) RegisterComponent(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.components[name] = &Component{
		Name:       name,
		Status:     StatusHealthy,
		LastCheck:  time.Now(),
		ErrorCount: 0,
	}
}

// RecordSuccess records a successful operation for a component
func (h *Checker) RecordSuccess(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if comp, exists := h.components[name]; exists {
		comp.SuccessCount++
		comp.LastCheck = time.Now()

		// If we had errors, gradually recover to healthy
		if comp.ErrorCount > 0 {
			comp.ErrorCount--
		}

		if comp.ErrorCount == 0 {
			comp.Status = StatusHealthy
			comp.Message = ""
		}
	}
}

// RecordError records an error for a component
func (h *Checker) RecordError(name string, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if comp, exists := h.components[name]; exists {
		comp.ErrorCount++
		comp.LastCheck = time.Now()
		comp.Message = err.Error()

		// Determine status based on error count
		if comp.ErrorCount >= 10 {
			comp.Status = StatusUnhealthy
		} else if comp.ErrorCount >= 3 {
			comp.Status = StatusDegraded
		}
	}
}

// GetComponentStatus returns the status of a specific component
func (h *Checker) GetComponentStatus(name string) *Component {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if comp, exists := h.components[name]; exists {
		// Return a copy to avoid race conditions
		compCopy := *comp
		return &compCopy
	}

	return nil
}

// GetOverallStatus returns the overall system health status
func (h *Checker) GetOverallStatus() Status {
	h.mu.RLock()
	defer h.mu.RUnlock()

	hasUnhealthy := false
	hasDegraded := false

	for _, comp := range h.components {
		switch comp.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusHealthy
}

// GetAllComponents returns a snapshot of all components
func (h *Checker) GetAllComponents() map[string]*Component {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]*Component, len(h.components))
	for name, comp := range h.components {
		compCopy := *comp
		result[name] = &compCopy
	}

	return result
}

// IsHealthy returns true if the system is healthy
func (h *Checker) IsHealthy() bool {
	return h.GetOverallStatus() == StatusHealthy
}
