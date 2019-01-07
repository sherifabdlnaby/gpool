package gpool

import (
	"context"
	"math"
	"sync"
)

// workerPool is an implementation of gpool.Pool interface to bound concurrency using a Worker goroutines.
type workerPool struct {
	workerCount     int
	workerPoolQueue chan *worker
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.Mutex
	status          uint8
}

// NewWorkerPool is an implementation of gpool.Pool interface to bound concurrency using a Semaphore.
func NewWorkerPool(workerCount int) *workerPool {
	newWorkerPool := workerPool{
		workerCount: workerCount,
		mu:          sync.Mutex{},
		status:      poolClosed,
	}

	// Cancel immediately - So that ErrPoolClosed will be returned by Enqueues
	// This to check for pool_closed pool without checking 'status' and avoid redundant if condition because ctx will be
	// checked anyway.
	// A Not Canceled context will be assigned by Start().
	newWorkerPool.ctx, newWorkerPool.cancel = context.WithCancel(context.TODO())
	newWorkerPool.cancel()

	return &newWorkerPool
}

// Start the Pool, otherwise it will not accept any job.
func (w *workerPool) Start() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.workerCount < 1 {
		return ErrPoolInvalidSize
	}

	if w.status == poolStarted {
		return nil
	}

	ctx := context.Background()

	// Init chans and Stuff
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.workerPoolQueue = make(chan *worker, w.workerCount)

	// Add and Spin Up N Workers
	for i := 0; i < w.workerCount; i++ {
		w.addNewWorker()
	}

	w.status = poolStarted

	return nil
}

// Stop the Pool.
// 1- ALL Blocked/Waiting jobs will return immediately.
// 2- All Jobs Processing will finish successfully
// 3- Stop() WILL Block until all running jobs is done.
func (w *workerPool) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.status == poolClosed {
		return
	}
	// stop every-worker
	for i := 0; i < w.workerCount; i++ {
		w.removeWorker()
	}

	// send cancellation signal ( to unblock blocked enqueues )
	w.cancel()

	// Wait for All Active working Jobs to end.
	w.wg.Wait()

	w.status = poolClosed
}

// Resize the pool size in concurrent-safe way
func (w *workerPool) Resize(newSize int) error {
	if newSize < 1 {
		return ErrPoolInvalidSize
	}

	// Resize
	w.mu.Lock()

	delta := newSize - w.workerCount

	// If already pool_started live resize it.
	if w.status == poolStarted && delta != 0 {
		if delta > 0 {
			for i := 0; i < delta; i++ {
				w.addNewWorker()
			}
		} else {
			delta = int(math.Abs(float64(delta)))
			for i := 0; i < delta; i++ {
				w.removeWorker()
			}
		}
	}

	w.workerCount = newSize

	w.mu.Unlock()

	return nil
}

// Enqueue Process job func(){} and returns ONCE the func has pool_started (not after it ends)
// If the pool is full pool.Enqueue() will block until either:
// 		1- A worker/slot in the pool is done and is ready to take another job.
//		2- The Job context is canceled.
//		3- The Pool is pool_closed by pool.Stop().
// @Returns nil once the job has pool_started.
// @Returns ErrPoolClosed if the pool is not running.
// @Returns ErrJobCanceled if the job Enqueued context was canceled before the job could be processed by the pool.
func (w *workerPool) Enqueue(ctx context.Context, job func()) error {
	select {
	// The Job was canceled through job's context, no need to DO the work now.
	case <-ctx.Done():
		return ErrJobCanceled
	// Pool Cancellation Signal.
	case <-w.ctx.Done():
		return ErrPoolClosed
	case workerReceiveChan := <-w.workerPoolQueue:
		workerReceiveChan.receive <- job
		return nil
	}
}

// TryEnqueue will not block if the pool is full, will return true once the job has pool_started processing or false if
// the pool is pool_closed or full.
func (w *workerPool) TryEnqueue(f func()) bool {
	select {
	case workerReceiveChan := <-w.workerPoolQueue:
		workerReceiveChan.receive <- f
		return true
	default:
		return false
	}
}

// GetSize return the current size of the pool.
func (w *workerPool) GetSize() int {
	return w.workerCount
}

func (w *workerPool) addNewWorker() {
	worker := worker{
		receive:    make(chan func()),
		workerPool: w.workerPoolQueue,
	}
	worker.ctx, worker.cancel = context.WithCancel(w.ctx)

	// Start worker and start consuming
	worker.Start(&w.wg)
}

func (w *workerPool) removeWorker() {
	// Pick worker from pool
	worker := <-w.workerPoolQueue

	// Cancel that particular worker
	worker.cancel()

	// Send a dummy job
	// Workers ONLY check if they're canceled when they're putting themselves in the ready pool.
	// Since we picked a worker from the pool, it is expecting a job before checking if they're canceled or not.
	// That's a better alternative than nested checking for same ctx in worker loop ? //TODO.
	worker.receive <- func() {}

	return
}

// --------- WORKER --------- ///

type worker struct {
	workerPool chan *worker
	receive    chan func()
	ctx        context.Context
	cancel     context.CancelFunc
}

func (w *worker) Start(wg *sync.WaitGroup) bool {
	wg.Add(1)

	// Send Signal that the below goroutine has pool_started already
	//  -->	Handles when TryEnqueue Returns FALSE if immediately called after starting the pool
	//  	As the worker goroutines may still have not yet launched
	started := make(chan bool, 1)

	go func() {
		defer wg.Done()
		started <- true
		for {
			select {
			// If Worker Canceled or Pool Canceled
			case <-w.ctx.Done():
				return
			default:
				// Add self to ready-pool
				w.workerPool <- w
				task := <-w.receive
				task()
			}
		}
	}()

	return <-started
}
