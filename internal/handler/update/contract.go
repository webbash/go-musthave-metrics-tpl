package update

import "context"

type metricsService interface {
	Update(ctx context.Context, metricType, metricName, metricValue string) error
}
