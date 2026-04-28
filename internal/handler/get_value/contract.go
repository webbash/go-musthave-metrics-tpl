package get_value

import "context"

type metricsService interface {
	Get(ctx context.Context, metricType, metricName string) (string, error)
}
