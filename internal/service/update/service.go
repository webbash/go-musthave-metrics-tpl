package update

type Service struct {
	storage storage
}

func New(storage storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) IncrementCounter(counterName string, value int64) error {
	s.storage.IncrementCounter(counterName, value)

	return nil
}

func (s *Service) UpdateGauge(counterName string, value float64) error {
	s.storage.UpdateGauge(counterName, value)

	return nil
}
