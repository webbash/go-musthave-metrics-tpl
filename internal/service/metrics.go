package service

import (
	"context"
	"errors"
	"strconv"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

var (
	ErrInvalidMetricValue = errors.New("invalid metric value")
	ErrMetricNotFound     = errors.New("metric not found")
	ErrUnknownMetricType  = errors.New("unknown metric type")
)

type MetricsRepository interface {
	GetAllGauges(ctx context.Context) map[string]float64
	GetAllCounters(ctx context.Context) map[string]int64
	GetCounter(ctx context.Context, metricName string) (int64, bool)
	GetGauge(ctx context.Context, metricName string) (float64, bool)
	IncrementCounter(ctx context.Context, metricName string, value int64)
	UpdateGauge(ctx context.Context, metricName string, value float64)
}

type MetricsService struct {
	repository MetricsRepository
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

func (s *MetricsService) UpdateMetric(ctx context.Context) map[string]float64 {
	// TODO ...
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
