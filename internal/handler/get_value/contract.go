package get_value

type storage interface {
	GetCounter(metricName string) (int64, bool)
	GetGauge(metricName string) (float64, bool)
}
