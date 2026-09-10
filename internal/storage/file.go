// Package storage provides file-backed metric storage.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type FileStorage struct {
	path string
}

func NewFileStorage(path string) *FileStorage {
	return &FileStorage{
		path: path,
	}
}

func (s *FileStorage) Save(metrics []models.Metrics) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("could not marshal metrics: %w", err)
	}

	err = os.MkdirAll(filepath.Dir(s.path), 0755)
	if err != nil {
		return fmt.Errorf("failed to create subdirs: %w", err)
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *FileStorage) Load() ([]models.Metrics, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read temp file: %w", err)
	}
	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, fmt.Errorf("failed to deserialize json: %w", err)
	}
	return metrics, nil
}
