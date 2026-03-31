package update

type storage interface {
	IncrementCounter(metricName string, value int64)
	UpdateGauge(metricName string, value float64)
}
