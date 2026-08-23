package audit

import (
	"go.uber.org/zap"
)

// Subject distributes audit events to its registered observers.
type Subject struct {
	observers []Observer
	logger    *zap.SugaredLogger
}

// NewSubject creates an audit event subject that reports observer errors to logger.
func NewSubject(logger *zap.SugaredLogger) *Subject {
	return &Subject{logger: logger}
}

// AddObserver registers an observer to receive subsequent audit events.
func (s *Subject) AddObserver(o Observer) {
	s.observers = append(s.observers, o)
}

// Notify sends an audit event to every registered observer.
// Observer errors are logged and do not prevent the remaining observers from running.
func (s *Subject) Notify(e Event) error {
	for _, o := range s.observers {
		err := o.Observe(e)
		if err != nil {
			s.logger.Errorf("error notifying observer: %v", err)
		}
	}
	return nil
}
