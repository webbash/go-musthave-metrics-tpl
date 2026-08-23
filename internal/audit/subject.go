package audit

import (
	"sync"

	"go.uber.org/zap"
)

// Subject distributes audit events to its registered observers.
type Subject struct {
	channels []chan Event
	closed   bool
	mu       *sync.RWMutex
	wg       *sync.WaitGroup
	logger   *zap.SugaredLogger
}

// NewSubject creates an audit event subject that reports observer errors to logger.
func NewSubject(logger *zap.SugaredLogger) *Subject {
	return &Subject{
		logger:   logger,
		channels: make([]chan Event, 0),
		mu:       &sync.RWMutex{},
		wg:       &sync.WaitGroup{},
	}
}

// AddObserver registers an observer to receive subsequent audit events.
func (s *Subject) AddObserver(o Observer) {
	s.channels = append(s.channels, s.addChannel(o))
}

func (s *Subject) addChannel(o Observer) chan Event {
	ch := make(chan Event, 5)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		for event := range ch {
			err := o.Observe(event)
			if err != nil {
				s.logger.Errorw("Failed to observe event", "err", err)
			}
		}
	}()

	return ch
}

// Notify sends an audit event to every registered observer.
// Observer errors are logged and do not prevent the remaining observers from running.
func (s *Subject) Notify(e Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return
	}

	for _, ch := range s.channels {
		select {
		case ch <- e:
		default:
			// channel is full, skip event
			s.logger.Warnw("Audit event channel is full, dropping event", "event", e)
		}
	}
	return
}

func (s *Subject) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true
	for _, ch := range s.channels {
		close(ch)
	}

	s.wg.Wait()
}
