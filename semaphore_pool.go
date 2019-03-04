package gpool

import (
	"context"
	"github.com/sherifabdlnaby/semaphore"
	"math"
	"sync"
)

// SemaphorePool is an implementation of gpool.Pool interface to bound concurrency using a Semaphore.
type SemaphorePool struct {
	workerCount int
	semaphore   *semaphore.Weighted
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
		semaphore:   semaphore.NewWeighted(math.MaxInt64),
		mu:          sync.Mutex{},
		status:      poolClosed,
	}

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
	w.semaphore = semaphore.NewWeighted(int64(w.workerCount))
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

	// Try to Acquire the whole Semaphore ( This will block until all ACTIVE works are done )
	// And also plays as a lock to change pool status.
	_ = w.semaphore.Acquire(context.Background(), int64(w.workerCount))

	w.status = poolClosed

	// Release the Semaphore so that subsequent enqueues will not block and return ErrPoolClosed.
	w.semaphore.Release(int64(w.workerCount))

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
		w.semaphore.Resize(int64(newSize))
	}

	w.mu.Unlock()

	return nil
}

// Enqueue Process job `func(){...}` and returns ONCE the func has started executing (not after it ends/return)
//
// If the pool is full `Enqueue()` will block until either:
// 		1- A worker/slot in the pool is done and is ready to take the job.
//		2- The Job context is canceled.
//		3- The Pool is closed by `pool.Stop()`.
// @Returns nil once the job has started executing.
// @Returns ErrPoolClosed if the pool is not running.
// @Returns ErrJobCanceled if the job Enqueued context was canceled before the job could be processed by the pool.
func (w *SemaphorePool) Enqueue(ctx context.Context, job func()) error {
	// Acquire 1 from semaphore ( aka Acquire one worker )
	err := w.semaphore.Acquire(ctx, 1)

	// The Job was canceled through job's context, no need to DO the work now.
	if err != nil {
		return ErrJobCanceled
	}

	// Check if pool is running
	// (This is safe as the semaphore in stop() will only change status if it acquired the full semaphore.).
	if w.status != poolStarted {
		return ErrPoolClosed
	}

	// Run the job and return.
	go func() {
		// Run the job
		job()

		w.semaphore.Release(1)
	}()

	return nil
}

// EnqueueAndWait Process job `func(){...}` and returns ONCE the func has returned.
//
// If the pool is full `Enqueue()` will block until either:
// 		1- A worker/slot in the pool is done and is ready to take the job.
//		2- The Job context is canceled.
//		3- The Pool is closed by `pool.Stop()`.
// @Returns nil once the job has executed and returned.
// @Returns ErrPoolClosed if the pool is not running.
// @Returns ErrJobCanceled if the job Enqueued context was canceled before the job could be processed by the pool.
func (w *SemaphorePool) EnqueueAndWait(ctx context.Context, job func()) error {

	// Acquire 1 from semaphore ( aka Acquire one worker )
	err := w.semaphore.Acquire(ctx, 1)

	// The Job was canceled through job's context, no need to DO the work now.
	if err != nil {
		return ErrJobCanceled
	}

	// Check if pool is running
	// (This is safe as the semaphore in stop() will only change status if it acquired the full semaphore.).
	if w.status != poolStarted {
		return ErrPoolClosed
	}

	// Run the job
	job()

	w.semaphore.Release(1)

	return nil
}

// TryEnqueue will not block if the pool is full, will return true once the job has started processing or false if the pool is closed or full.
func (w *SemaphorePool) TryEnqueue(job func()) bool {
	// Acquire 1 from semaphore ( aka Acquire one worker )
	// False if semaphore is full or status not started.
	if !w.semaphore.TryAcquire(1) || w.status != poolStarted {
		return false
	}

	// Start Job
	go func() {
		// Run the Function
		job()

		w.semaphore.Release(1)
	}()

	return true
}

// TryEnqueue will not block if the pool is full, will return true once the job has finished processing or false if the pool is closed or full.
func (w *SemaphorePool) TryEnqueueAndWait(job func()) bool {
	// Acquire 1 from semaphore ( aka Acquire one worker )
	// False if semaphore is full or status not started.
	if !w.semaphore.TryAcquire(1) || w.status != poolStarted {
		return false
	}

	// Run the Function
	job()

	w.semaphore.Release(1)

	return true
}

// GetSize returns the maximum size of the pool.
func (w *SemaphorePool) GetSize() int {
	return w.workerCount
}

// GetCurrent returns the current size of the pool.
func (w *SemaphorePool) GetCurrent() int {
	return int(w.semaphore.Current())
}

// GetWaiting return the current size of jobs waiting in the pool.
func (w *SemaphorePool) GetWaiting() int {
	return int(w.semaphore.Waiters())
}
