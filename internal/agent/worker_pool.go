package agent

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type WorkerPool struct {
	sender  *Sender
	batches <-chan Batch
	workers int
	wg      sync.WaitGroup
	logger  *zap.SugaredLogger
}

func NewWorkerPool(sender *Sender, workers int, batches <-chan Batch, logger *zap.SugaredLogger) *WorkerPool {
	return &WorkerPool{
		sender:  sender,
		batches: batches,
		workers: workers,
		logger:  logger,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.workers; i++ {
		workerID := i
		wp.wg.Add(1)
		go func(id int) {
			defer func() {
				wp.logger.Infow(
					"worker stopped",
					"worker", id,
				)
				wp.wg.Done()
			}()
			for {
				select {
				case <-ctx.Done():
					wp.logger.Warnw("Stopping worker", "id", id)
					return
				case batch, ok := <-wp.batches:
					if !ok {
						wp.logger.Warnw("Stopping worker", "id", id)
						return
					}
					if err := wp.sender.Send(ctx, batch.Metrics); err != nil {
						wp.logger.Errorw("Failed to send request by worker", "id", id)
					}
				}
			}
		}(workerID)
	}
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}
