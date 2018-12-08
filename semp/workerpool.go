package semp

import (
	"context"
	"errors"
	"golang.org/x/sync/semaphore"
)

type WorkerPoolSemaphore struct {
	WorkerCount int64
	semaphore   semaphore.Weighted
	ctx         context.Context
	cancel      context.CancelFunc
}

func (w *WorkerPoolSemaphore) Start() {
	return
}

func (w *WorkerPoolSemaphore) Stop() {
	w.ctx.Done()
	_ = w.semaphore.Acquire(context.TODO(), w.WorkerCount)
	return
}

func (w *WorkerPoolSemaphore) Enqueue(f func()) error {
	err := w.semaphore.Acquire(w.ctx, 1)
	if err != nil {
		return ErrWorkerPoolClosed2
	}
	go func() {
		f()
		w.semaphore.Release(1)
	}()
	return nil

}

var (
	//DIFFERENT NAME FOR DEBUGGING ONLY //TODO only1
	ErrWorkerPoolClosed1 = errors.New("pool is closed ( By checking for OK )")
	ErrWorkerPoolClosed2 = errors.New("pool is closed ( By checking <-ctx.Done() )")
	ErrWorkTimeout       = errors.New("timeout")
)

func NewSempWorker(workerCount int64) *WorkerPoolSemaphore {
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
