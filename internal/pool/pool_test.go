package pool

import (
	"sync"
	"sync/atomic"
	"testing"
)

type testObject struct {
	value     string
	resetting *atomic.Int32
}

func (o *testObject) Reset() {
	o.value = ""
	if o.resetting != nil {
		o.resetting.Add(1)
	}
}

func TestPoolPutCallsReset(t *testing.T) {
	var resets atomic.Int32
	p := New[*testObject]()
	object := &testObject{value: "temporary value", resetting: &resets}

	p.Put(object)

	if got := resets.Load(); got != 1 {
		t.Fatalf("Reset called %d times, want 1", got)
	}
	if object.value != "" {
		t.Fatalf("object was not reset: %q", object.value)
	}
}

func TestPoolGetReturnsPutObject(t *testing.T) {
	p := New[*testObject]()
	object := &testObject{value: "value"}

	p.Put(object)
	got := p.Get()

	if got != object {
		t.Fatal("Get returned a different object")
	}
	if got.value != "" {
		t.Fatalf("object was not reset: %q", got.value)
	}
}

func TestPoolGetReturnsLastPutObjectFirst(t *testing.T) {
	p := New[*testObject]()
	first := &testObject{}
	second := &testObject{}

	p.Put(first)
	p.Put(second)

	if got := p.Get(); got != second {
		t.Fatal("Get did not return the last put object")
	}
	if got := p.Get(); got != first {
		t.Fatal("Get did not return the previous object")
	}
}

func TestPoolConcurrentUse(t *testing.T) {
	const workers = 100

	p := New[*testObject]()
	for i := 0; i < workers; i++ {
		p.Put(&testObject{})
	}

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			object := p.Get()
			if object == nil {
				return
			}
			object.value = "temporary value"
			p.Put(object)
		}()
	}

	wg.Wait()
}
