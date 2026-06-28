package repository

import (
	"context"
	"fmt"

	"github.com/webbash/go-musthave-metrics-tpl.git/internal/storage"
)

type FileRepository struct {
	*MemStorage
	storage *storage.FileStorage
}

func NewFileRepository(
	memStorage *MemStorage,
	storage *storage.FileStorage,
) *FileRepository {
	return &FileRepository{
		MemStorage: memStorage,
		storage:    storage,
	}
}

func (r *FileRepository) IncrementCounter(
	ctx context.Context,
	metricName string,
	value int64,
) error {
	err := r.MemStorage.IncrementCounter(ctx, metricName, value)
	if err != nil {
		return fmt.Errorf("file storage: called increment counter: %w", err)
	}

	if err := r.persist(); err != nil {
		return fmt.Errorf("file storage: persist: %w", err)
	}

	return nil
}

func (r *FileRepository) UpdateGauge(
	ctx context.Context,
	metricName string,
	value float64,
) error {
	err := r.MemStorage.UpdateGauge(ctx, metricName, value)
	if err != nil {
		return fmt.Errorf("file storage: called increment counter: %w", err)
	}

	if err := r.persist(); err != nil {
		return fmt.Errorf("file storage: persist: %w", err)
	}

	return nil
}

func (r *FileRepository) persist() error {
	metrics, err := r.MemStorage.GetAllMetrics(context.TODO())
	if err != nil {
		return fmt.Errorf("get all metrics: %w", err)
	}

	if err := r.storage.Save(metrics); err != nil {
		return fmt.Errorf("save metrics to file: %w", err)
	}

	return nil
}
