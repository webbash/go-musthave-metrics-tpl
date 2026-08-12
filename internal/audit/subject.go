package audit

import (
	"go.uber.org/zap"
)

type Subject struct {
	observers []Observer
	logger    *zap.SugaredLogger
}

func NewSubject(logger *zap.SugaredLogger) *Subject {
	return &Subject{logger: logger}
}

func (s *Subject) AddObserver(o Observer) {
	s.observers = append(s.observers, o)
}

func (s *Subject) Notify(e Event) error {
	for _, o := range s.observers {
		err := o.Observe(e)
		if err != nil {
			s.logger.Errorf("error notifying observer: %v", err)
		}
	}
	return nil
}
