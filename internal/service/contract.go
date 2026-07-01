package service

import (
	"context"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type MetricsRepository interface {
	GetAllGauges(ctx context.Context) (map[string]float64, error)
	GetAllCounters(ctx context.Context) (map[string]int64, error)
	GetCounter(ctx context.Context, metricName string) (int64, error)
	GetGauge(ctx context.Context, metricName string) (float64, error)
	IncrementCounter(ctx context.Context, metricName string, value int64) error
	UpdateGauge(ctx context.Context, metricName string, value float64) error
	GetAllMetrics(ctx context.Context) ([]models.Metrics, error)
	UpdateBatch(ctx context.Context, metrics []models.Metrics) error
}

type MetricsFileStorage interface {
	Save(metrics []models.Metrics) error
}
