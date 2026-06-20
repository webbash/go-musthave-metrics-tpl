package storage

import models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"

type MockFileStorage struct {
	saveCalled bool
	savedData  []models.Metrics
}

func (m *MockFileStorage) Save(metrics []models.Metrics) error {
	m.saveCalled = true
	m.savedData = metrics
	return nil
}

func (m *MockFileStorage) Load() ([]models.Metrics, error) {
	return nil, nil
}
