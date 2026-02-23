package update

import models "github.com/webbash/go-musthave-metrics-tpl.git/internal/model"

type storage interface {
	Put(metrics models.Metrics)
}
