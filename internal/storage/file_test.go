package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

func TestFileStorage_SaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "metrics.json")
	storage := NewFileStorage(path)

	value := 23.5
	delta := int64(4)
	metrics := []models.Metrics{
		{ID: "temperature", MType: models.Gauge, Value: &value},
		{ID: "requests", MType: models.Counter, Delta: &delta},
	}

	require.NoError(t, storage.Save(metrics))
	loaded, err := storage.Load()
	require.NoError(t, err)
	assert.Equal(t, metrics, loaded)
}

func TestFileStorage_LoadErrors(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		storage := NewFileStorage(filepath.Join(t.TempDir(), "missing.json"))
		_, err := storage.Load()
		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "metrics.json")
		require.NoError(t, os.WriteFile(path, []byte("not json"), 0644))

		_, err := NewFileStorage(path).Load()
		assert.Error(t, err)
	})
}
