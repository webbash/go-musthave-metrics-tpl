package storage

import (
	"fmt"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
	"os"
)

type FileStorage struct {
	path string
}

func NewFileStorage(path string) (*FileStorage, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0755)

	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s", err)
	}

	return &FileStorage{
		path: path,
	}, nil
}

func (s *FileStorage) Save(metrics []models.Metrics) error {

}

func (s *FileStorage) Load() ([]models.Metrics, error) {}
