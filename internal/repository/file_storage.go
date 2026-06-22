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
) {
	r.MemStorage.IncrementCounter(ctx, metricName, value)

	if err := r.persist(); err != nil {

		// либо логируем ошибку
		// либо меняем сигнатуры методов и возвращаем её наверх
	}
}

func (r *FileRepository) UpdateGauge(
	ctx context.Context,
	metricName string,
	value float64,
) {
	r.MemStorage.UpdateGauge(ctx, metricName, value)

	if err := r.persist(); err != nil {
		// либо логируем ошибку
		// либо меняем сигнатуры методов и возвращаем её наверх
	}
}

func (r *FileRepository) persist() error {
	metrics := r.MemStorage.GetAllMetrics()

	if err := r.storage.Save(metrics); err != nil {
		return fmt.Errorf("save metrics to file: %w", err)
	}

	return nil
}
