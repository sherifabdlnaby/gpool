package gpool

import (
	"context"
	"errors"
	"math"
	"runtime"
	"sync"

	"github.com/sherifabdlnaby/semaphore"
)

var (
	// ErrPoolInvalidSize Returned if the Size of pool < 1.
	ErrPoolInvalidSize = errors.New("pool size is invalid, pool size must be >= 0")

	// ErrPoolClosed Error Returned if the Pool has not pool_started yet, or was stopped.
	ErrPoolClosed = errors.New("pool is closed")
)

const (
	poolStopped = iota
	poolStarted
)

// Pool is an implementation of gpool.Pool interface to bound concurrency using a Semaphore.
type Pool struct {
	workerCount int
	semaphore   *semaphore.Weighted
	mu          sync.Mutex
	status      uint8
}

// NewPool returns a pool that uses semaphore implementation.
// Returns ErrPoolInvalidSize if size is < 1.
func NewPool(size int) *Pool {

	if size < 0 {
		panic(ErrPoolInvalidSize)
		return nil
	}

	// If size is zero, set default == no. of cpus
	if size == 0 {
		size = runtime.NumCPU()
	}

	newWorkerPool := Pool{
		workerCount: size,
		semaphore:   semaphore.NewWeighted(math.MaxInt64),
		mu:          sync.Mutex{},
		status:      poolStopped,
	}

	// start pool.
	newWorkerPool.Start()

	return &newWorkerPool
}

// Start the Pool, otherwise it will not accept any job.
//
// Subsequent calls to Start will not have any effect unless Stop() is called.
func (w *Pool) Start() {
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
func (w *Pool) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.status == poolStopped {
		return
	}

	// Try to Acquire the whole Semaphore ( This will block until all ACTIVE works are done )
	// And also plays as a lock to change pool status.
	_ = w.semaphore.Acquire(context.Background(), int64(w.workerCount))

	w.status = poolStopped

	// Release the Semaphore so that subsequent enqueues will not block and return ErrPoolClosed.
	w.semaphore.Release(int64(w.workerCount))

	return
}

// Resize the pool size in concurrent-safe way.
//
//  `Resize` can enlarge the pool and any blocked enqueue will unblock after pool is resized, in case of shrinking the pool `resize` will not affect any already processing job.
func (w *Pool) Resize(newSize int) {
	if newSize < 1 {
		panic(ErrPoolInvalidSize)
	}

	// If newSize is zero, set default == no. of cpus
	if newSize == 0 {
		newSize = runtime.NumCPU()
	}

	// Resize
	w.mu.Lock()

	w.workerCount = newSize

	// If already pool_started live resize semaphore limit.
	if w.status == poolStarted {
		w.semaphore.Resize(int64(w.workerCount))
	}

	w.mu.Unlock()
}

// Enqueue Process job `func(){...}` and returns ONCE the func has started executing (not after it ends/return)
//
// If the pool is full `Enqueue()` will block until either:
// 		1- A worker/slot in the pool is done and is ready to take the job.
//		2- The Job context is canceled.
//		3- The Pool is closed by `pool.Stop()`.
// @Returns nil once the job has started executing.
// @Returns ErrPoolClosed if the pool is not running.
// @Returns ctx.Err() if the job Enqueued context was canceled before the job could be processed by the pool.
func (w *Pool) Enqueue(ctx context.Context, job func()) error {
	// Acquire 1 from semaphore ( aka Acquire one worker )
	err := w.semaphore.Acquire(ctx, 1)

	// The Job was canceled through job's context, no need to DO the work now.
	if err != nil {
		return err
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
// @Returns ctx.Err() if the job Enqueued context was canceled before the job could be processed by the pool.
func (w *Pool) EnqueueAndWait(ctx context.Context, job func()) error {

	// Acquire 1 from semaphore ( aka Acquire one worker )
	err := w.semaphore.Acquire(ctx, 1)

	// The Job was canceled through job's context, no need to DO the work now.
	if err != nil {
		return err
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
func (w *Pool) TryEnqueue(job func()) bool {
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

// TryEnqueueAndWait will not block if the pool is full, will return true once the job has finished processing or false if the pool is closed or full.
func (w *Pool) TryEnqueueAndWait(job func()) bool {
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
func (w *Pool) GetSize() int {
	return w.workerCount
}

// GetCurrent returns the current size of the pool.
func (w *Pool) GetCurrent() int {
	return int(w.semaphore.Current())
}

// GetWaiting return the current size of jobs waiting in the pool.
func (w *Pool) GetWaiting() int {
	return int(w.semaphore.Waiters())
}
