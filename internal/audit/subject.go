package audit

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// Subject distributes audit events to its registered observers.
type Subject struct {
	channels []chan Event
	closed   bool
	ctx      context.Context
	mu       *sync.RWMutex
	wg       *sync.WaitGroup
	logger   *zap.SugaredLogger
}

// NewSubject creates an audit event subject with the given lifecycle context
// that reports observer errors to logger.
func NewSubject(ctx context.Context, logger *zap.SugaredLogger) *Subject {
	return &Subject{
		logger:   logger,
		ctx:      ctx,
		channels: make([]chan Event, 0),
		mu:       &sync.RWMutex{},
		wg:       &sync.WaitGroup{},
	}
}

// AddObserver registers an observer and starts a worker that processes events
// from the observer's queue.
func (s *Subject) AddObserver(o Observer) {
	s.channels = append(s.channels, s.addChannel(o))
}

func (s *Subject) addChannel(o Observer) chan Event {
	ch := make(chan Event, 5)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		for event := range ch {
			err := o.Observe(s.ctx, event)
			if err != nil {
				s.logger.Errorw("Failed to observe event", "err", err)
			}
		}
	}()

	return ch
}

// Notify sends an event to each observer's queue without blocking.
// The event is dropped for an observer if its queue is full.
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

// Close stops accepting new events, closes all observer queues, and waits for
// their buffered events to be processed.
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
