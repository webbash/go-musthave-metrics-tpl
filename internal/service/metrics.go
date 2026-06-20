package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

var (
	ErrInvalidMetricValue = errors.New("invalid metric value")
	ErrMetricNotFound     = errors.New("metric not found")
	ErrUnknownMetricType  = errors.New("unknown metric type")
	ErrInvalidMetric      = errors.New("metric is invalid")
)

type MetricsRepository interface {
	GetAllGauges(ctx context.Context) map[string]float64
	GetAllCounters(ctx context.Context) map[string]int64
	GetCounter(ctx context.Context, metricName string) (int64, bool)
	GetGauge(ctx context.Context, metricName string) (float64, bool)
	IncrementCounter(ctx context.Context, metricName string, value int64)
	UpdateGauge(ctx context.Context, metricName string, value float64)
	GetAllMetrics() []models.Metrics
}

type MetricsFileStorage interface {
	Save(metrics []models.Metrics) error
}

type MetricsService struct {
	repository    MetricsRepository
	fileStorage   MetricsFileStorage
	storeInterval int
}

func NewMetricsService(repository MetricsRepository, fileStorage MetricsFileStorage, storeInterval int) *MetricsService {
	return &MetricsService{
		repository:    repository,
		fileStorage:   fileStorage,
		storeInterval: storeInterval,
	}
}

func (s *MetricsService) Update(ctx context.Context, metricType, metricName, metricValue string) error {
	switch metricType {
	case models.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return ErrInvalidMetricValue
		}

		s.repository.IncrementCounter(ctx, metricName, value)
		return nil
	case models.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return ErrInvalidMetricValue
		}

		s.repository.UpdateGauge(ctx, metricName, value)
		return nil
	default:
		return ErrUnknownMetricType
	}
}

func (s *MetricsService) UpdateMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	switch metric.MType {
	case models.Counter:
		if metric.Delta == nil {
			return models.Metrics{}, ErrInvalidMetricValue
		}

		s.repository.IncrementCounter(ctx, metric.ID, *metric.Delta)
		counterValue, _ := s.repository.GetCounter(ctx, metric.ID)

		if s.storeInterval == 0 {
			err := s.fileStorage.Save(s.repository.GetAllMetrics())
			if err != nil {
				return models.Metrics{}, fmt.Errorf("failed to save metrics to file: %w", err)
			}
		}

		metric.Delta = &counterValue

		return metric, nil
	case models.Gauge:
		if metric.Value == nil {
			return models.Metrics{}, ErrInvalidMetricValue
		}

		s.repository.UpdateGauge(ctx, metric.ID, *metric.Value)

		if s.storeInterval == 0 {
			err := s.fileStorage.Save(s.repository.GetAllMetrics())
			if err != nil {
				return models.Metrics{}, fmt.Errorf("failed to save metrics to file: %w", err)
			}
		}

		return metric, nil
	default:
		return models.Metrics{}, ErrUnknownMetricType
	}

}

func (s *MetricsService) GetMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	if metric.ID == "" || metric.MType == "" {
		return models.Metrics{}, ErrInvalidMetric
	}

	switch metric.MType {
	case models.Gauge:
		value, ok := s.repository.GetGauge(ctx, metric.ID)
		if !ok {
			return models.Metrics{}, ErrMetricNotFound
		}

		metric.Value = &value
		return metric, nil

	case models.Counter:
		value, ok := s.repository.GetCounter(ctx, metric.ID)
		if !ok {
			return models.Metrics{}, ErrMetricNotFound
		}
		metric.Delta = &value
		return metric, nil
	default:
		return models.Metrics{}, ErrUnknownMetricType
	}
}

func (s *MetricsService) Get(ctx context.Context, metricType, metricName string) (string, error) {
	switch metricType {
	case models.Counter:
		value, ok := s.repository.GetCounter(ctx, metricName)
		if !ok {
			return "", ErrMetricNotFound
		}

		return strconv.FormatInt(value, 10), nil
	case models.Gauge:
		value, ok := s.repository.GetGauge(ctx, metricName)
		if !ok {
			return "", ErrMetricNotFound
		}

		return strconv.FormatFloat(value, 'f', -1, 64), nil
	default:
		return "", ErrUnknownMetricType
	}
}

func (s *MetricsService) GetAll(ctx context.Context) (map[string]float64, map[string]int64) {
	return s.repository.GetAllGauges(ctx), s.repository.GetAllCounters(ctx)
}
