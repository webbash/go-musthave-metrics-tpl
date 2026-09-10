package getvaluelist

import (
	"context"
)

type metricsRepository interface {
	GetAllGauges(ctx context.Context) (map[string]float64, error)
	GetAllCounters(ctx context.Context) (map[string]int64, error)
}
