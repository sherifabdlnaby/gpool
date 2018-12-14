package dpool

import (
	"context"
	"golang.org/x/sync/semaphore"
)

type SemaphorePool struct {
	WorkerCount int64
	semaphore   semaphore.Weighted
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewSemaphorePool(workerCount int64) *SemaphorePool {
	newWorkerPool := SemaphorePool{
		WorkerCount: workerCount,
	}
	return &newWorkerPool
}

func (w *SemaphorePool) Start() {
	ctx := context.Background()
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.semaphore = *semaphore.NewWeighted(w.WorkerCount)
	return
}

func (w *SemaphorePool) Stop() {
	// Send Cancellation Signal to stop all waiting work
	w.cancel()

	// Try to Acquire the whole Semaphore ( This will block until all ACTIVE works are done )
	_ = w.semaphore.Acquire(context.TODO(), w.WorkerCount)

	// Release the Semaphore so that subsequent
	w.semaphore.Release(w.WorkerCount)

	return
}

func (w *SemaphorePool) Enqueue(ctx context.Context, f func()) error {
	// Acquire 1 from semaphore ( aka Acquire one worker )
	err := w.semaphore.Acquire(ctx, 1)

	// The Job was canceled through job's context, no need to DO the work now.
	if err != nil {
		return ErrWorkerPoolClosed2
	}

	select {
	// Pool Cancellation Signal
	case <-w.ctx.Done():
		w.semaphore.Release(1)
		return ErrWorkerPoolClosed1
	default:
		go func() {
			defer func() {
				w.semaphore.Release(1)
				/*				if r := recover(); r != nil {
								fmt.Println("Recovered in f", r)
							}*/
			}()

			// Run the Function
			f()
		}()
	}

	return nil
}

func (w *SemaphorePool) TryEnqueue(f func()) (bool, error) {
	// Acquire 1 from semaphore ( aka Acquire one worker )
	if !w.semaphore.TryAcquire(1) {
		return false, nil
	}

	go func() {
		defer func() {
			w.semaphore.Release(1)
			/*				if r := recover(); r != nil {
							fmt.Println("Recovered in f", r)
						}*/
		}()

		// Run the Function
		f()
	}()

	return true, nil
}
