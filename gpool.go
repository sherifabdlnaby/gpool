package gpool

import (
	"context"
	"errors"
)

// Pool Manages a pool of goroutines to bound concurrency, A Job is Enqueued to the pool and only N jobs can be processed
// concurrently, pool.Start() Initialize the pool, If the pool is full pool.Enqueue() will block until either:
//		1- A worker/slot in the pool is done and is ready to take another job.
//		2- The Job context is canceled.
//		3- The Pool is pool_closed by pool.Stop().
type Pool interface {
	// Start the Pool, otherwise it will not accept any job.
	//
	// Subsequent calls to Start will not have any effect unless Stop() is called. Will return ErrPoolInvalidSize
	// If pool size is < 1.
	Start() error

	// Stop the Pool.
	//
	//		1- ALL Blocked/Waiting jobs will return immediately.
	// 		2- All Jobs Processing will finish successfully
	//		3- Stop() WILL Block until all running jobs is done.
	// Subsequent Calls to Stop() will have no effect unless start() is called.
	Stop()

	// Resize the pool size in concurrent-safe way.
	//
	//  `Resize` can enlarge the pool and any blocked enqueue will unblock after pool is resized, in case of shrinking the pool `resize` will not affect any already processing job.
	Resize(int) error

	// Enqueue Process job func(){} and returns ONCE the func has pool_started (not after it ends)
	//
	// If the pool is full pool.Enqueue() will block until either:
	// 		1- A worker/slot in the pool is done and is ready to take another job.
	//		2- The Job context is canceled.
	//		3- The Pool is pool_closed by pool.Stop().
	// @Returns nil once the job has pool_started.
	// @Returns ErrPoolClosed if the pool is not running.
	// @Returns ErrJobCanceled if the job Enqueued context was canceled before the job could be processed by the pool.
	Enqueue(context.Context, func()) error

	// TryEnqueue will not block if the pool is full, will return true once the job has started processing or false if the pool is closed or full.
	TryEnqueue(func()) bool

	// GetSize return the current size of the pool.
	GetSize() int
}

var (
	// ErrPoolInvalidSize Returned if the Size of pool < 1.
	ErrPoolInvalidSize = errors.New("pool size is invalid, pool size must be > 0")

	// ErrPoolClosed Error Returned if the Pool has not pool_started yet, or was stopped.
	ErrPoolClosed = errors.New("pool is closed")

	// ErrJobCanceled Error Returned if the job's context was canceled while blocking waiting for the pool.
	ErrJobCanceled = errors.New("job canceled")
)

const (
	poolClosed = iota
	poolStarted
)
