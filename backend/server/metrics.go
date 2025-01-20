package server

import (
	"sync"
)

type Metrics struct {
	activeSources   map[string]bool
	alertsTriggered int
	mu              sync.Mutex
}

func NewMetrics() *Metrics {
	return &Metrics{activeSources: make(map[string]bool)}
}

// IncrementSource marks a source as active
func (m *Metrics) IncrementSource(source string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.activeSources[source] = true
}

// GetActiveSources returns the number of active sources
func (m *Metrics) GetActiveSources() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.activeSources)
}

// IncrementAlerts increments the alerts triggered count
func (m *Metrics) IncrementAlerts() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alertsTriggered++
}

// GetAlertsTriggered returns the number of alerts triggered
func (m *Metrics) GetAlertsTriggered() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.alertsTriggered
}
