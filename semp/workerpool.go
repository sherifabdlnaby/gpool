package semp

import (
	"context"
	"errors"
	"golang.org/x/sync/semaphore"
	"pipeline/work"
)

type WorkerPoolSemaphore struct {
	WorkerCount int
	semaphore   semaphore.Weighted
	ctx         context.Context
	cancel      context.CancelFunc
}

func (w *WorkerPoolSemaphore) Start() {
	return
}

func (w *WorkerPoolSemaphore) Stop() {
	w.ctx.Done()
	return
}

func (w *WorkerPoolSemaphore) Enqueue(work worker.Work, payload int) (<-chan int, error) {
	err := w.semaphore.Acquire(w.ctx, 1)
	if err != nil {
		return nil, ErrWorkerPoolClosed2
	}
	resultChan := make(chan int, 1)
	go func() {
		resultChan <- work.Run(payload)
		w.semaphore.Release(1)
	}()
	return resultChan, nil

}

var (
	//DIFFERENT NAME FOR DEBUGGING ONLY //TODO only1
	ErrWorkerPoolClosed1 = errors.New("pool is closed ( By checking for OK )")
	ErrWorkerPoolClosed2 = errors.New("pool is closed ( By checking <-ctx.Done() )")
	ErrWorkTimeout       = errors.New("timeout")
)

func NewSempWorker(workerCount int) *WorkerPoolSemaphore {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	newWorkerPool := WorkerPoolSemaphore{
		WorkerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
		semaphore:   *semaphore.NewWeighted(int64(workerCount)),
	}
	return &newWorkerPool
}
