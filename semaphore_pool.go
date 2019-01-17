package gpool

import (
	"context"
	"github.com/marusama/semaphore"
	"sync"
)

// SemaphorePool is an implementation of gpool.Pool interface to bound concurrency using a Semaphore.
type SemaphorePool struct {
	workerCount int
	semaphore   semaphore.Semaphore
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	status      uint8
}

// NewSemaphorePool returns a pool that uses semaphore implementation.
// Returns ErrPoolInvalidSize if size is < 1.
func NewSemaphorePool(size int) (Pool, error) {

	if size < 1 {
		return nil, ErrPoolInvalidSize
	}

	newWorkerPool := SemaphorePool{
		workerCount: size,
		semaphore:   semaphore.New(1),
		mu:          sync.Mutex{},
		status:      poolClosed,
	}

	// Cancel immediately - So that ErrPoolClosed will be returned by Enqueues
	// This to check for pool_closed pool without checking 'status' and avoid redundant if condition because ctx will be
	// checked anyway.
	// A Not Canceled context will be assigned by Start().
	newWorkerPool.ctx, newWorkerPool.cancel = context.WithCancel(context.TODO())
	newWorkerPool.cancel()

	return &newWorkerPool, nil
}

// Start the Pool, otherwise it will not accept any job.
//
// Subsequent calls to Start will not have any effect unless Stop() is called.
func (w *SemaphorePool) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.status == poolStarted {
		return
	}

	w.status = poolStarted
	ctx := context.Background()
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.semaphore = semaphore.New(w.workerCount)
}

// Stop the Pool.
//
//		1- ALL Blocked/Waiting jobs will return immediately.
// 		2- All Jobs Processing will finish successfully
//		3- Stop() WILL Block until all running jobs is done.
// Subsequent Calls to Stop() will have no effect unless start() is called.
func (w *SemaphorePool) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.status == poolClosed {
		return
	}

	// Send Cancellation Signal to stop all waiting work
	w.cancel()

	/*
		// Try to Acquire the whole Semaphore ( This will block until all ACTIVE works are done )
		_ = w.semaphore.Acquire(context.Background(), w.workerCount)
	*/

	// Try to Acquire the whole Semaphore ( This will block until all ACTIVE works are done )
	// Acquire 1 by 1 because a bug I found when N > GOMAXPROCS -> https://github.com/cockroachdb/cockroach/issues/33554
	for i := 0; i < w.workerCount; i++ {
		_ = w.semaphore.Acquire(context.TODO(), 1)
	}

	// Release the Semaphore so that subsequent enqueues will not block and return ErrPoolClosed.
	w.semaphore.Release(w.workerCount)

	w.status = poolClosed

	return
}

// Resize the pool size in concurrent-safe way.
//
//  `Resize` can enlarge the pool and any blocked enqueue will unblock after pool is resized, in case of shrinking the pool `resize` will not affect any already processing job.
func (w *SemaphorePool) Resize(newSize int) error {
	if newSize < 1 {
		return ErrPoolInvalidSize
	}

	// Resize
	w.mu.Lock()

	w.workerCount = newSize

	// If already pool_started live resize semaphore limit.
	if w.status == poolStarted {
		w.semaphore.SetLimit(newSize)
	}

	w.mu.Unlock()

	return nil
}

// Enqueue Process job func(){} and returns ONCE the func has pool_started (not after it ends)
//
// If the pool is full pool.Enqueue() will block until either:
// 		1- A worker/slot in the pool is done and is ready to take another job.
//		2- The Job context is canceled.
//		3- The Pool is pool_closed by pool.Stop().
// @Returns nil once the job has pool_started.
// @Returns ErrPoolClosed if the pool is not running.
// @Returns ErrJobCanceled if the job Enqueued context was canceled before the job could be processed by the pool.
func (w *SemaphorePool) Enqueue(ctx context.Context, job func()) error {
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

// TryEnqueue will not block if the pool is full, will return true once the job has started processing or false if the pool is closed or full.
func (w *SemaphorePool) TryEnqueue(job func()) bool {
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

// GetSize return the current size of the pool.
func (w *SemaphorePool) GetSize() int {
	return w.workerCount
}
