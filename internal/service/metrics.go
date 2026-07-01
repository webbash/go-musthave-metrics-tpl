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

type MetricsService struct {
	repository    MetricsRepository
	fileStorage   MetricsFileStorage
	storeInterval int
}

func NewMetricsService(repository MetricsRepository) *MetricsService {
	return &MetricsService{
		repository: repository,
	}
}

func (s *MetricsService) Update(ctx context.Context, metricType, metricName, metricValue string) error {
	switch metricType {
	case models.Counter:
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return ErrInvalidMetricValue
		}

		errInc := s.repository.IncrementCounter(ctx, metricName, value)
		if errInc != nil {
			return fmt.Errorf("failed to increment counter: %w", errInc)
		}
		return nil
	case models.Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return ErrInvalidMetricValue
		}

		errUpd := s.repository.UpdateGauge(ctx, metricName, value)
		if errUpd != nil {
			return fmt.Errorf("failed to update gauge: %w", errUpd)
		}
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

		err := s.repository.IncrementCounter(ctx, metric.ID, *metric.Delta)
		if err != nil {
			return models.Metrics{}, fmt.Errorf("failed to increment counter: %w", err)
		}
		counterValue, _ := s.repository.GetCounter(ctx, metric.ID)

		metric.Delta = &counterValue

		return metric, nil
	case models.Gauge:
		if metric.Value == nil {
			return models.Metrics{}, ErrInvalidMetricValue
		}

		err := s.repository.UpdateGauge(ctx, metric.ID, *metric.Value)
		if err != nil {
			return models.Metrics{}, fmt.Errorf("failed to update gauge: %w", err)
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
		value, err := s.repository.GetGauge(ctx, metric.ID)
		if err != nil {
			return models.Metrics{}, ErrMetricNotFound
		}

		metric.Value = &value
		return metric, nil

	case models.Counter:
		value, err := s.repository.GetCounter(ctx, metric.ID)
		if err != nil {
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
		value, err := s.repository.GetCounter(ctx, metricName)
		if err != nil {
			return "", ErrMetricNotFound
		}

		return strconv.FormatInt(value, 10), nil
	case models.Gauge:
		value, err := s.repository.GetGauge(ctx, metricName)
		if err != nil {
			return "", ErrMetricNotFound
		}

		return strconv.FormatFloat(value, 'f', -1, 64), nil
	default:
		return "", ErrUnknownMetricType
	}
}

func (s *MetricsService) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	for _, metric := range metrics {
		if metric.MType != models.Counter && metric.MType != models.Gauge {
			return ErrUnknownMetricType
		}
		if metric.MType == models.Counter && metric.Delta == nil {
			return ErrInvalidMetricValue
		}
		if metric.MType == models.Gauge && metric.Value == nil {
			return ErrInvalidMetricValue
		}
	}

	err := s.repository.UpdateBatch(ctx, metrics)
	if err != nil {
		return fmt.Errorf("failed to update batch metrics: %w", err)
	}

	return nil
}
