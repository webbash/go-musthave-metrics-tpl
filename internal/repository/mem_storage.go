package repository

import (
	"context"
	"fmt"
	"sync"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type MemStorage struct {
	mu      sync.RWMutex
	counter map[string]int64
	gauge   map[string]float64
}

func NewMemStorageFromMetrics(metrics []models.Metrics) *MemStorage {
	counter := make(map[string]int64)
	gauge := make(map[string]float64)

	for _, metric := range metrics {
		if metric.MType == models.Counter {
			counter[metric.ID] = *metric.Delta
		}
		if metric.MType == models.Gauge {
			gauge[metric.ID] = *metric.Value
		}
	}

	return &MemStorage{
		counter: counter,
		gauge:   gauge,
	}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counter: make(map[string]int64),
		gauge:   make(map[string]float64),
	}
}

func (m *MemStorage) GetAllGauges(_ context.Context) (map[string]float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]float64, len(m.gauge))
	for name, val := range m.gauge {
		result[name] = val
	}
	return result, nil
}

func (m *MemStorage) GetAllCounters(_ context.Context) (map[string]int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]int64, len(m.counter))
	for name, val := range m.counter {
		result[name] = val
	}
	return result, nil
}

func (m *MemStorage) GetCounter(_ context.Context, metricName string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.counter[metricName]
	if !ok {
		return 0, fmt.Errorf("counter %s not found", metricName)
	}
	return value, nil
}

func (m *MemStorage) GetGauge(_ context.Context, metricName string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.gauge[metricName]
	if !ok {
		return 0, fmt.Errorf("gauge %s not found", metricName)
	}
	return value, nil
}

func (m *MemStorage) IncrementCounter(_ context.Context, metricName string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter[metricName] += value

	return nil
}

func (m *MemStorage) UpdateGauge(_ context.Context, metricName string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gauge[metricName] = value

	return nil
}

func (m *MemStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	counter, err := m.GetAllCounters(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get all counters: %w", err)
	}
	result := make([]models.Metrics, 0)

	for name, value := range counter {
		result = append(result, models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &value,
		})
	}

	gauges, err := m.GetAllGauges(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get all gauges: %w", err)
	}

	for name, value := range gauges {
		result = append(result, models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
		})
	}

	return result, nil
}

func (m *MemStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case models.Counter:
			err := m.IncrementCounter(ctx, metric.ID, *metric.Delta)
			if err != nil {
				return fmt.Errorf("failed to increment counter %s: %w", metric.ID, err)
			}
		case models.Gauge:
			err := m.UpdateGauge(ctx, metric.ID, *metric.Value)
			if err != nil {
				return fmt.Errorf("failed to update gauge %s: %w", metric.ID, err)
			}
		}
	}

	return nil
}
