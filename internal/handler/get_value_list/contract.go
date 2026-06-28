package get_value_list

import (
	"context"
)

type repository interface {
	GetAllGauges(ctx context.Context) (map[string]float64, error)
	GetAllCounters(ctx context.Context) (map[string]int64, error)
}
