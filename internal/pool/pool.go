// Package pool provides a reusable pool for resettable objects.
package pool

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	objects []T
	mu      sync.Mutex
}

func New[T Resettable]() *Pool[T] {
	return &Pool[T]{}
}

func (p *Pool[T]) Put(obj T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	obj.Reset()
	p.objects = append(p.objects, obj)
}

func (p *Pool[T]) Get() T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.objects) == 0 {
		var zero T
		return zero
	}

	obj := p.objects[len(p.objects)-1]
	p.objects = p.objects[:len(p.objects)-1]

	return obj
}
