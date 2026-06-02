package repository

import (
	"context"
	"sync"
)

type MemStorage struct {
	mu      sync.RWMutex
	counter map[string]int64
	gauge   map[string]float64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counter: make(map[string]int64),
		gauge:   make(map[string]float64),
	}
}

func (m *MemStorage) GetAllGauges(_ context.Context) map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]float64, len(m.gauge))
	for name, val := range m.gauge {
		result[name] = val
	}
	return result
}

func (m *MemStorage) GetAllCounters(_ context.Context) map[string]int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]int64, len(m.counter))
	for name, val := range m.counter {
		result[name] = val
	}
	return result
}

func (m *MemStorage) GetCounter(_ context.Context, metricName string) (int64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.counter[metricName]
	return value, ok
}

func (m *MemStorage) GetGauge(_ context.Context, metricName string) (float64, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.gauge[metricName]
	return value, ok
}

func (m *MemStorage) IncrementCounter(_ context.Context, metricName string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter[metricName] += value
}

func (m *MemStorage) UpdateGauge(_ context.Context, metricName string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauge[metricName] = value
}
