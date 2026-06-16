package update_v2

import "context"

type metricsService interface {
	Update(ctx context.Context, metricType, metricName, metricValue string) error
}
