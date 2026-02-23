package repository

import models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"

type MemStorage struct {
	metrics map[string]map[string]models.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]map[string]models.Metrics),
	}
}

func (m *MemStorage) Put(metrics models.Metrics) {
	m.metrics[metrics.MType][metrics.ID] = metrics
}
