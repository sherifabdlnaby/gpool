package gpool

import (
	"context"
	"golang.org/x/sync/semaphore"
)

type SemaphorePool struct {
	WorkerCount int
	semaphore   semaphore.Weighted
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewSemaphorePool(workerCount int) *SemaphorePool {
	newWorkerPool := SemaphorePool{
		WorkerCount: workerCount,
		semaphore:   *semaphore.NewWeighted(1),
	}

	// Cancel immediately - So that ErrPoolClosed will be returned by Enqueues
	// A Not Canceled context will be assigned by Start().
	newWorkerPool.ctx, newWorkerPool.cancel = context.WithCancel(context.TODO())
	newWorkerPool.cancel()

	return &newWorkerPool
}

func (w *SemaphorePool) Start() {
	ctx := context.Background()
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.semaphore = *semaphore.NewWeighted(int64(w.WorkerCount))
	return
}

func (w *SemaphorePool) Stop() {
	// Send Cancellation Signal to stop all waiting work
	w.cancel()

	// Try to Acquire the whole Semaphore ( This will block until all ACTIVE works are done )
	_ = w.semaphore.Acquire(context.TODO(), int64(w.WorkerCount))

	// Release the Semaphore so that subsequent enqueues will not block and return ErrPoolClosed.
	w.semaphore.Release(int64(w.WorkerCount))

	// This to give a breathing space for Enqueue to not wait for a Semaphore of 0 size.
	// so that Enqueue won't block ( will still send ErrPoolClosed )
	if w.WorkerCount == 0 {
		w.semaphore = *semaphore.NewWeighted(1)
	}

	return
}

func (w *SemaphorePool) Enqueue(ctx context.Context, job func()) error {
	// Acquire 1 from semaphore ( aka Acquire one worker )
	err := w.semaphore.Acquire(ctx, 1)

	// The Job was canceled through job's context, no need to DO the work now.
	if err != nil {
		return ErrJobTimeout
	}

	select {
	// Pool Cancellation Signal
	case <-w.ctx.Done():
		w.semaphore.Release(1)
		return ErrPoolClosed
	default:
		go func() {
			defer func() {
				w.semaphore.Release(1)
				/*				if r := recover(); r != nil {
								fmt.Println("Recovered in job", r)
							}*/
			}()

			// Run the Function
			job()
		}()
	}

	return nil
}

func (w *SemaphorePool) TryEnqueue(job func()) bool {
	// Acquire 1 from semaphore ( aka Acquire one worker )
	if !w.semaphore.TryAcquire(1) {
		return false
	}

	go func() {
		defer func() {
			w.semaphore.Release(1)
			/*				if r := recover(); r != nil {
							fmt.Println("Recovered in job", r)
						}*/
		}()

		// Run the Function
		job()
	}()

	return true
}
