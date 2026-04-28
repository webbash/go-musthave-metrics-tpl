package get_value_list

import (
	"context"
)

type metricsService interface {
	GetAll(ctx context.Context) (map[string]float64, map[string]int64)
}
