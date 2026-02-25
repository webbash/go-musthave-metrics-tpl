package update

type service interface {
	IncrementCounter(counterName string, value int64) error
	UpdateGauge(counterName string, value float64) error
}
