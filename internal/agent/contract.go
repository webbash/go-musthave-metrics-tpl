package agent

import (
	"context"

	models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"
)

// MetricsSender delivers one collected batch using the configured transport.
type MetricsSender interface {
	Send(context.Context, []models.Metrics) error
}
