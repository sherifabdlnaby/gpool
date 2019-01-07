package gpool

import (
	"context"
	"sync"
)

// workerPool is an implementation of gpool.Pool interface to bound concurrency using a Worker goroutines.
type workerPool struct {
	workerCount int
	workerQueue chan chan func()
	workers     []worker
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewWorkerPool is an implementation of gpool.Pool interface to bound concurrency using a Semaphore.
func NewWorkerPool(workerCount int) *workerPool {
	newWorkerPool := workerPool{
		workerCount: workerCount,
	}

	// Cancel immediately - So that ErrPoolClosed will be returned by Enqueues
	// A Not Canceled context will be assigned by Start().
	newWorkerPool.ctx, newWorkerPool.cancel = context.WithCancel(context.TODO())
	newWorkerPool.cancel()

	return &newWorkerPool
}

// Start the Pool, otherwise it will not accept any job.
func (w *workerPool) Start() error {
	if w.workerCount < 1 {
		return ErrPoolInvalidSize
	}

	ctx := context.Background()

	// Init chans and Stuff
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.workerQueue = make(chan chan func(), w.workerCount)
	w.workers = make([]worker, w.workerCount)

	// Spin Up Workers
	for i := 0; i < w.workerCount; i++ {

		worker := worker{
			ID:      i,
			Receive: make(chan func()),
			Worker:  w.workerQueue,
		}

		// Start worker and start consuming
		worker.Start(w.ctx, &w.wg)

		// Store workers
		w.workers = append(w.workers, worker)
	}

	return nil
}

// Stop the Pool.
// 1- ALL Blocked/Waiting jobs will return immediately.
// 2- All Jobs Processing will finish successfully
// 3- Stop() WILL Block until all running jobs is done.
func (w *workerPool) Stop() {
	// Send Cancellation Signal to stop all waiting work
	w.cancel()

	// Wait for All Active working Jobs to end.
	w.wg.Wait()
}

// Enqueue Process job func(){} and returns ONCE the func has pool_started (not after it ends)
// If the pool is full pool.Enqueue() will block until either:
// 		1- A worker/slot in the pool is done and is ready to take another job.
//		2- The Job context is canceled.
//		3- The Pool is pool_closed by pool.Stop().
// @Returns nil once the job has pool_started.
// @Returns ErrPoolClosed if the pool is not running.
// @Returns ErrJobCanceled if the job Enqueued context was canceled before the job could be processed by the pool.
func (w *workerPool) Enqueue(ctx context.Context, f func()) error {
	select {
	// The Job was canceled through job's context, no need to DO the work now.
	case <-ctx.Done():
		return ErrJobCanceled
	// Pool Cancellation Signal.
	case <-w.ctx.Done():
		return ErrPoolClosed
	case workerReceiveChan := <-w.workerQueue:
		select {
		// Send the job to worker.
		case workerReceiveChan <- f:
			return nil
		// This is in-case the worker has been stopped (via cancellation signal) BEFORE we send the job to it,
		// Hence it won't receive the job on workerReceiveChan and will block.
		default:
			return ErrPoolClosed
		}
	}
}

// TryEnqueue will not block if the pool is full, will return true once the job has pool_started processing or false if
// the pool is pool_closed or full.
func (w *workerPool) TryEnqueue(f func()) bool {
	select {
	case workerReceiveChan := <-w.workerQueue:
		select {
		// Send the job to worker.
		case workerReceiveChan <- f:
			return true
		// This is in-case the worker has been stopped (via cancellation signal) BEFORE we send the job to it,
		// Hence it won't receive the job and would block.
		default:
			return false
		}
	default:
		return false
	}
}

// --------- WORKER --------- ///
type worker struct {
	ID      int
	Worker  chan chan func()
	Receive chan func()
}

func (w *worker) Start(ctx context.Context, wg *sync.WaitGroup) bool {
	wg.Add(1)

	// Send Signal that the below goroutine has pool_started already
	//  -->	Handles when TryEnqueue Returns FALSE if immediately called after starting the pool
	//  	As the worker goroutines may still have not yet launched
	started := make(chan bool, 1)

	go func() {
		defer wg.Done()
		started <- true
		for {
			w.Worker <- w.Receive
			select {
			case task := <-w.Receive:
				task()
			case <-ctx.Done():
				return
			}
		}
	}()

	return <-started
}
