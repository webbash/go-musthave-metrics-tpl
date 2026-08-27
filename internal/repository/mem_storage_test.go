package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	"github.com/webbash/go-musthave-metrics-tpl.git/internal/storage"
)

func TestMemStorage(t *testing.T) {
	ctx := context.Background()
	value := 23.5
	delta := int64(4)
	metrics := []models.Metrics{
		{ID: "temperature", MType: models.Gauge, Value: &value},
		{ID: "requests", MType: models.Counter, Delta: &delta},
	}
	mem := NewMemStorageFromMetrics(metrics)

	gauges, err := mem.GetAllGauges(ctx)
	require.NoError(t, err)
	assert.Equal(t, map[string]float64{"temperature": value}, gauges)
	counters, err := mem.GetAllCounters(ctx)
	require.NoError(t, err)
	assert.Equal(t, map[string]int64{"requests": delta}, counters)

	assert.Equal(t, value, func() float64 { v, _ := mem.GetGauge(ctx, "temperature"); return v }())
	assert.Equal(t, delta, func() int64 { v, _ := mem.GetCounter(ctx, "requests"); return v }())
	assert.Error(t, func() error { _, getErr := mem.GetGauge(ctx, "missing"); return getErr }())
	assert.Error(t, func() error { _, getErr := mem.GetCounter(ctx, "missing"); return getErr }())

	require.NoError(t, mem.UpdateGauge(ctx, "load", 0.75))
	require.NoError(t, mem.IncrementCounter(ctx, "requests", 2))
	updatedCounter, err := mem.GetCounter(ctx, "requests")
	require.NoError(t, err)
	assert.Equal(t, int64(6), updatedCounter)

	all, err := mem.GetAllMetrics(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	newValue := 1.25
	newDelta := int64(3)
	require.NoError(t, mem.UpdateMany(ctx, []models.Metrics{
		{ID: "load", MType: models.Gauge, Value: &newValue},
		{ID: "new_requests", MType: models.Counter, Delta: &newDelta},
	}))
	assert.Equal(t, newValue, func() float64 { v, _ := mem.GetGauge(ctx, "load"); return v }())
	assert.Equal(t, newDelta, func() int64 { v, _ := mem.GetCounter(ctx, "new_requests"); return v }())
}

func TestFileRepositoryPersistsUpdates(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "metrics.json")
	repo := NewFileRepository(NewMemStorage(), storage.NewFileStorage(path))

	require.NoError(t, repo.UpdateGauge(ctx, "temperature", 23.5))
	require.NoError(t, repo.IncrementCounter(ctx, "requests", 1))

	loaded, err := storage.NewFileStorage(path).Load()
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
}
