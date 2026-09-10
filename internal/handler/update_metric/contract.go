package updatemetric

import (
	"context"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type metricsService interface {
	UpdateMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error)
}
