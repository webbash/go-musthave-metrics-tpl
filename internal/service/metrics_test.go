package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/repository"
)

type repositoryStub struct {
	getCounterFn       func(context.Context, string) (int64, error)
	getGaugeFn         func(context.Context, string) (float64, error)
	incrementCounterFn func(context.Context, string, int64) error
	updateGaugeFn      func(context.Context, string, float64) error
	updateManyFn       func(context.Context, []models.Metrics) error
}

func (r repositoryStub) GetAllGauges(context.Context) (map[string]float64, error) {
	return nil, nil
}

func (r repositoryStub) GetAllCounters(context.Context) (map[string]int64, error) {
	return nil, nil
}

func (r repositoryStub) GetCounter(ctx context.Context, name string) (int64, error) {
	if r.getCounterFn != nil {
		return r.getCounterFn(ctx, name)
	}
	return 0, nil
}

func (r repositoryStub) GetGauge(ctx context.Context, name string) (float64, error) {
	if r.getGaugeFn != nil {
		return r.getGaugeFn(ctx, name)
	}
	return 0, nil
}

func (r repositoryStub) IncrementCounter(ctx context.Context, name string, value int64) error {
	if r.incrementCounterFn != nil {
		return r.incrementCounterFn(ctx, name, value)
	}
	return nil
}

func (r repositoryStub) UpdateGauge(ctx context.Context, name string, value float64) error {
	if r.updateGaugeFn != nil {
		return r.updateGaugeFn(ctx, name, value)
	}
	return nil
}

func (r repositoryStub) GetAllMetrics(context.Context) ([]models.Metrics, error) {
	return nil, nil
}

func (r repositoryStub) UpdateMany(ctx context.Context, metrics []models.Metrics) error {
	if r.updateManyFn != nil {
		return r.updateManyFn(ctx, metrics)
	}
	return nil
}

func TestMetricsService_Update(t *testing.T) {
	ctx := context.Background()
	storage := repository.NewMemStorage()
	svc := NewMetricsService(storage)

	t.Run("updates counter", func(t *testing.T) {
		err := svc.Update(ctx, models.Counter, "requests", "3")
		require.NoError(t, err)

		value, err := storage.GetCounter(ctx, "requests")
		require.NoError(t, err)
		assert.Equal(t, int64(3), value)
	})

	t.Run("updates gauge", func(t *testing.T) {
		err := svc.Update(ctx, models.Gauge, "temperature", "23.5")
		require.NoError(t, err)

		value, err := storage.GetGauge(ctx, "temperature")
		require.NoError(t, err)
		assert.Equal(t, 23.5, value)
	})

	tests := []struct {
		name       string
		metricType string
		value      string
		want       error
	}{
		{name: "invalid counter", metricType: models.Counter, value: "3.5", want: ErrInvalidMetricValue},
		{name: "invalid gauge", metricType: models.Gauge, value: "hot", want: ErrInvalidMetricValue},
		{name: "unknown type", metricType: "histogram", value: "1", want: ErrUnknownMetricType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, svc.Update(ctx, tt.metricType, "metric", tt.value), tt.want)
		})
	}

	dbErr := errors.New("storage unavailable")
	t.Run("wraps counter repository error", func(t *testing.T) {
		svc := NewMetricsService(repositoryStub{
			incrementCounterFn: func(context.Context, string, int64) error { return dbErr },
		})
		err := svc.Update(ctx, models.Counter, "requests", "1")
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("wraps gauge repository error", func(t *testing.T) {
		svc := NewMetricsService(repositoryStub{
			updateGaugeFn: func(context.Context, string, float64) error { return dbErr },
		})
		err := svc.Update(ctx, models.Gauge, "temperature", "23.5")
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestMetricsService_UpdateMetric(t *testing.T) {
	ctx := context.Background()
	storage := repository.NewMemStorage()
	svc := NewMetricsService(storage)

	delta := int64(2)
	updatedCounter, err := svc.UpdateMetric(ctx, models.Metrics{ID: "requests", MType: models.Counter, Delta: &delta})
	require.NoError(t, err)
	require.NotNil(t, updatedCounter.Delta)
	assert.Equal(t, int64(2), *updatedCounter.Delta)

	value := 23.5
	updatedGauge, err := svc.UpdateMetric(ctx, models.Metrics{ID: "temperature", MType: models.Gauge, Value: &value})
	require.NoError(t, err)
	require.NotNil(t, updatedGauge.Value)
	assert.Equal(t, 23.5, *updatedGauge.Value)

	tests := []struct {
		name   string
		metric models.Metrics
		want   error
	}{
		{name: "counter without delta", metric: models.Metrics{ID: "requests", MType: models.Counter}, want: ErrInvalidMetricValue},
		{name: "gauge without value", metric: models.Metrics{ID: "temperature", MType: models.Gauge}, want: ErrInvalidMetricValue},
		{name: "unknown type", metric: models.Metrics{ID: "metric", MType: "histogram"}, want: ErrUnknownMetricType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.UpdateMetric(ctx, tt.metric)
			assert.ErrorIs(t, err, tt.want)
		})
	}

	dbErr := errors.New("storage unavailable")
	t.Run("wraps counter repository error", func(t *testing.T) {
		svc := NewMetricsService(repositoryStub{
			incrementCounterFn: func(context.Context, string, int64) error { return dbErr },
		})
		_, err := svc.UpdateMetric(ctx, models.Metrics{ID: "requests", MType: models.Counter, Delta: &delta})
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("wraps gauge repository error", func(t *testing.T) {
		svc := NewMetricsService(repositoryStub{
			updateGaugeFn: func(context.Context, string, float64) error { return dbErr },
		})
		_, err := svc.UpdateMetric(ctx, models.Metrics{ID: "temperature", MType: models.Gauge, Value: &value})
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestMetricsService_GetMetric(t *testing.T) {
	ctx := context.Background()
	storage := repository.NewMemStorage()
	svc := NewMetricsService(storage)

	value := 23.5
	err := storage.UpdateGauge(ctx, "temperature", value)
	require.NoError(t, err)
	err = storage.IncrementCounter(ctx, "requests", 4)
	require.NoError(t, err)

	gauge, err := svc.GetMetric(ctx, models.Metrics{ID: "temperature", MType: models.Gauge})
	require.NoError(t, err)
	require.NotNil(t, gauge.Value)
	assert.Equal(t, value, *gauge.Value)

	counter, err := svc.GetMetric(ctx, models.Metrics{ID: "requests", MType: models.Counter})
	require.NoError(t, err)
	require.NotNil(t, counter.Delta)
	assert.Equal(t, int64(4), *counter.Delta)

	tests := []struct {
		name   string
		metric models.Metrics
		want   error
	}{
		{name: "missing id", metric: models.Metrics{MType: models.Gauge}, want: ErrInvalidMetric},
		{name: "missing type", metric: models.Metrics{ID: "temperature"}, want: ErrInvalidMetric},
		{name: "missing gauge", metric: models.Metrics{ID: "missing", MType: models.Gauge}, want: ErrMetricNotFound},
		{name: "missing counter", metric: models.Metrics{ID: "missing", MType: models.Counter}, want: ErrMetricNotFound},
		{name: "unknown type", metric: models.Metrics{ID: "metric", MType: "histogram"}, want: ErrUnknownMetricType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetMetric(ctx, tt.metric)
			assert.ErrorIs(t, err, tt.want)
		})
	}
}

func TestMetricsService_Get(t *testing.T) {
	ctx := context.Background()
	storage := repository.NewMemStorage()
	svc := NewMetricsService(storage)

	require.NoError(t, storage.UpdateGauge(ctx, "temperature", 23.5))
	require.NoError(t, storage.IncrementCounter(ctx, "requests", 4))

	value, err := svc.Get(ctx, models.Gauge, "temperature")
	require.NoError(t, err)
	assert.Equal(t, "23.5", value)

	value, err = svc.Get(ctx, models.Counter, "requests")
	require.NoError(t, err)
	assert.Equal(t, "4", value)

	assert.ErrorIs(t, func() error { _, err := svc.Get(ctx, models.Gauge, "missing"); return err }(), ErrMetricNotFound)
	assert.ErrorIs(t, func() error { _, err := svc.Get(ctx, models.Counter, "missing"); return err }(), ErrMetricNotFound)
	assert.ErrorIs(t, func() error { _, err := svc.Get(ctx, "histogram", "metric"); return err }(), ErrUnknownMetricType)
}

func TestMetricsService_UpdateMany(t *testing.T) {
	ctx := context.Background()
	storage := repository.NewMemStorage()
	svc := NewMetricsService(storage)

	value := 23.5
	delta := int64(1)
	err := svc.UpdateMany(ctx, []models.Metrics{
		{ID: "temperature", MType: models.Gauge, Value: &value},
		{ID: "requests", MType: models.Counter, Delta: &delta},
	})
	require.NoError(t, err)

	storedGauge, err := storage.GetGauge(ctx, "temperature")
	require.NoError(t, err)
	assert.Equal(t, value, storedGauge)
	storedCounter, err := storage.GetCounter(ctx, "requests")
	require.NoError(t, err)
	assert.Equal(t, delta, storedCounter)

	tests := []struct {
		name    string
		metrics []models.Metrics
		want    error
	}{
		{name: "unknown type", metrics: []models.Metrics{{ID: "metric", MType: "histogram"}}, want: ErrUnknownMetricType},
		{name: "counter without delta", metrics: []models.Metrics{{ID: "requests", MType: models.Counter}}, want: ErrInvalidMetricValue},
		{name: "gauge without value", metrics: []models.Metrics{{ID: "temperature", MType: models.Gauge}}, want: ErrInvalidMetricValue},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ErrorIs(t, svc.UpdateMany(ctx, tt.metrics), tt.want)
		})
	}

	dbErr := errors.New("storage unavailable")
	svc = NewMetricsService(repositoryStub{
		updateManyFn: func(context.Context, []models.Metrics) error { return dbErr },
	})
	assert.ErrorIs(t, svc.UpdateMany(ctx, []models.Metrics{{ID: "metric", MType: models.Gauge, Value: &value}}), dbErr)
}
