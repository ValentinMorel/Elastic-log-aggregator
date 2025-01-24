package server

import (
	"sync"
	"sync/atomic"
	"time"
)

type MetricsData struct {
	activeSources   sync.Map
	alertsTriggered atomic.Int32
}

func NewMetrics() *MetricsData {
	metrics := &MetricsData{activeSources: sync.Map{}}
	metrics.MonitorSources()
	return metrics

}

func (m *MetricsData) IncrementSource(source string) {
	m.activeSources.Store(source, time.Now())
}

func (m *MetricsData) GetActiveSources() int {

	count := 0
	m.activeSources.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func (m *MetricsData) IncrementAlerts() {
	m.alertsTriggered.Add(1)
}

func (m *MetricsData) GetAlertsTriggered() int {
	return int(m.alertsTriggered.Load())
}

func (m *MetricsData) MonitorSources() {
	go func() {
		for {
			//log.Printf("Active sources: %d", m.GetActiveSources())
			m.activeSources.Range(func(key, value any) bool {
				if time.Since(value.(time.Time)) > 10*time.Second {
					m.activeSources.Delete(key)
					time.Sleep(2 * time.Second)
				}
				return true
			})
			time.Sleep(2 * time.Second)
		}
	}()
}
