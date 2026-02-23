package update

type Service struct {
	storage storage
}

func New(storage storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Put(typeName, name string, value float64) error {

}
