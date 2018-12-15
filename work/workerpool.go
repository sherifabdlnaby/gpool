package dpool

import (
	"context"
	"sync"
)

// --------- POOL --------- ///

type WorkerPool struct {
	workerCount int
	workerQueue chan chan func()
	workers     []worker
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewWorkerPool(workerCount int) *WorkerPool {
	newWorkerPool := WorkerPool{
		workerCount: workerCount,
	}

	// Cancel immediately - So that ErrPoolClosed will be returned by Enqueues
	// A Not Canceled context will be assigned by Start().
	newWorkerPool.ctx, newWorkerPool.cancel = context.WithCancel(context.TODO())
	newWorkerPool.cancel()

	return &newWorkerPool
}

func (w *WorkerPool) Start() {
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
}

func (w *WorkerPool) Stop() {
	// Send Cancellation Signal to stop all waiting work
	w.cancel()

	// Wait for All Active working Jobs to end.
	w.wg.Wait()
}

func (w *WorkerPool) Enqueue(ctx context.Context, f func()) error {

	select {
	// The Job was canceled through job's context, no need to DO the work now.
	case <-ctx.Done():
		return ErrJobTimeout
	// Pool Cancellation Signal.
	case <-w.ctx.Done():
		return ErrPoolClosed
	case workerReceiveChan := <-w.workerQueue:
		select {
		// Send the job to worker.
		case workerReceiveChan <- f:
			return nil
		// This is in-case the worker has been stopped (via cancellation signal) BEFORE we send the job to it,
		// Hence it won't receive the job and would block.
		case <-w.ctx.Done():
			return ErrPoolClosed
		}
	}
}

func (w *WorkerPool) TryEnqueue(f func()) bool {
	select {

	case workerReceiveChan := <-w.workerQueue:
		select {
		// Send the job to worker.
		case workerReceiveChan <- f:
			return true
		// This is in-case the worker has been stopped (via cancellation signal) BEFORE we send the job to it,
		// Hence it won't receive the job and would block.
		case <-w.ctx.Done():
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

func (w *worker) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
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
}
