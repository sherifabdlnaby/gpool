package sempchan

import (
	"context"
	"errors"
)

type WorkerPoolSemaphore struct {
	WorkerCount int64
	semaphore   chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
}

func (w *WorkerPoolSemaphore) Start() {
	return
}

func (w *WorkerPoolSemaphore) Stop() {
	//log.Println("STOPPING...")
	w.cancel()
	for i := 0; i < int(w.WorkerCount); i++ {
		w.semaphore <- struct{}{}
	}
	for i := 0; i < int(w.WorkerCount); i++ {
		<-w.semaphore
	}
	return
}

func (w *WorkerPoolSemaphore) Enqueue(ctx context.Context, f func()) error {
	select {
	case <-w.ctx.Done():
		return ErrWorkerPoolClosed1
	case <-ctx.Done():
		return ErrWorkerPoolClosed2
	case w.semaphore <- struct{}{}:
		go func() {
			f()
			<-w.semaphore
		}()
	}

	return nil
}

var (
	//DIFFERENT NAME FOR DEBUGGING ONLY //TODO only1
	ErrWorkerPoolClosed1 = errors.New("pool is closed ( By checking w.ctx.Done() )")
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
		semaphore:   make(chan struct{}, workerCount),
	}
	return &newWorkerPool
}
