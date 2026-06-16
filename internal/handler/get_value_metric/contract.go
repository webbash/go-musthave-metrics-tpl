package get_value_metric

import (
	"context"
	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

type metricsService interface {
	GetMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error)
}
