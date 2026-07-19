package agent

import (
	"context"
	"log"
	"sync"
)

type WorkerPool struct {
	sender  *Sender
	batches <-chan Batch
	workers int
	wg      sync.WaitGroup
}

type Result struct {
	WorkerID int
	Err      error
}

func NewWorkerPool(sender *Sender, workers int, batches <-chan Batch) *WorkerPool {
	return &WorkerPool{
		sender:  sender,
		batches: batches,
		workers: workers,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) chan Result {
	resultCh := make(chan Result)

	for i := 0; i < wp.workers; i++ {
		workerID := i
		wp.wg.Add(1)
		go func(id int) {
			defer wp.wg.Done()
			for {
				select {
				case <-ctx.Done():
					log.Printf("Stopping worker %d", id)
					return
				case batch, ok := <-wp.batches:
					if !ok {
						return
					}
					err := wp.sender.Send(ctx, batch.Metrics)
					resultCh <- Result{WorkerID: id, Err: err}
				}
			}
		}(workerID)
	}

	go func() {
		wp.wg.Wait()
		close(resultCh)
	}()

	return resultCh
}
