package update_batch

import (
	"context"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type metricsService interface {
	UpdateMany(ctx context.Context, metrics []models.Metrics) error
}
