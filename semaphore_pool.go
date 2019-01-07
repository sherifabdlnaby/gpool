package gpool

import (
	"context"
	"github.com/marusama/semaphore"
)

// semaphorePool is an implementation of gpool.Pool interface to bound concurrency using a Semaphore.
type semaphorePool struct {
	workerCount int
	semaphore   semaphore.Semaphore
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewSemaphorePool is semaphorePool Constructor
func NewSemaphorePool(size int) *semaphorePool {
	newWorkerPool := semaphorePool{
		workerCount: size,
		semaphore:   semaphore.New(1),
	}

	// Cancel immediately - So that ErrPoolClosed will be returned by Enqueues
	// A Not Canceled context will be assigned by Start().
	newWorkerPool.ctx, newWorkerPool.cancel = context.WithCancel(context.TODO())
	newWorkerPool.cancel()

	return &newWorkerPool
}

// Start the Pool, otherwise it will not accept any job.
func (w *semaphorePool) Start() error {
	if w.workerCount < 1 {
		return ErrPoolInvalidSize
	}
	ctx := context.Background()
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.semaphore = semaphore.New(w.workerCount)
	return nil
}

// Stop the Pool.
// 1- ALL Blocked/Waiting jobs will return immediately.
// 2- All Jobs Processing will finish successfully
// 3- Stop() WILL Block until all running jobs is done.
func (w *semaphorePool) Stop() {
	// Send Cancellation Signal to stop all waiting work
	w.cancel()

	// Try to Acquire the whole Semaphore ( This will block until all ACTIVE works are done )
	_ = w.semaphore.Acquire(context.TODO(), w.workerCount)

	// Release the Semaphore so that subsequent enqueues will not block and return ErrPoolClosed.
	w.semaphore.Release(w.workerCount)

	return
}

// Enqueue Process job func(){} and returns ONCE the func has started (not after it ends)
// If the pool is full pool.Enqueue() will block until either:
// 		1- A worker/slot in the pool is done and is ready to take another job.
//		2- The Job context is canceled.
//		3- The Pool is closed by pool.Stop().
// @Returns nil once the job has started.
// @Returns ErrPoolClosed if the pool is not running.
// @Returns ErrJobCanceled if the job Enqueued context was canceled before the job could be processed by the pool.
func (w *semaphorePool) Enqueue(ctx context.Context, job func()) error {
	// Acquire 1 from semaphore ( aka Acquire one worker )
	err := w.semaphore.Acquire(ctx, 1)

	// The Job was canceled through job's context, no need to DO the work now.
	if err != nil {
		return ErrJobCanceled
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
				/*
					if r := recover(); r != nil {
					fmt.Println("Recovered in job", r)
					}*/
			}()

			// Run the job
			job()
		}()
	}

	return nil
}

// TryEnqueue will not block if the pool is full, will return true once the job has started processing or false if
// the pool is closed or full.
func (w *semaphorePool) TryEnqueue(job func()) bool {
	// Acquire 1 from semaphore ( aka Acquire one worker )
	if !w.semaphore.TryAcquire(1) {
		return false
	}

	// Start Job
	go func() {
		defer func() {
			w.semaphore.Release(1)
			/*
				if r := recover(); r != nil {
				fmt.Println("Recovered in a job", r)
				}*/
		}()
		// Run the Function
		job()
	}()

	return true
}
