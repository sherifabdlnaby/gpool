package workerpool

import (
	"context"
	"errors"
	"pipeline/work"
	"sync"
	"time"
)

type WorkerPool struct {
	WorkerCount int
	workerQueue chan chan worker.Task
	workers     []worker.Worker
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

var (
	//DIFFERENT NAME FOR DEBUGGING ONLY //TODO only1
	ErrWorkerPoolClosed1 = errors.New("pool is closed ( By checking for OK )")
	ErrWorkerPoolClosed2 = errors.New("pool is closed ( By checking <-ctx.Done() )")
	ErrWorkTimeout       = errors.New("timeout")
)

func NewWorkerPool(workerCount int) *WorkerPool {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	newWorkerPool := WorkerPool{
		WorkerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
		workerQueue: make(chan chan worker.Task, workerCount),
		workers:     make([]worker.Worker, workerCount),
	}
	return &newWorkerPool
}

func (w *WorkerPool) Start() {
	// Spin Up Workers
	for i := 0; i < w.WorkerCount; i++ {

		//log.Println(fmt.Sprintf("Starting worker [%d]...", i))

		workerx := worker.Worker{
			ID:      i,
			Receive: make(chan worker.Task),
			Worker:  w.workerQueue,
		}

		// Start worker and start consuming
		workerx.Start(w.ctx, &w.wg)

		// Store workers
		w.workers = append(w.workers, workerx)
	}
}

func (w *WorkerPool) Stop() {
	w.cancel()
	w.wg.Wait()
	close(w.workerQueue)
	// drain the queue
	for range w.workerQueue {

	}
}

func (w *WorkerPool) Enqueue(work worker.Work, payload int) (<-chan int, error) {
	resultChan := make(chan int, 1)

	// Check If Worker Pool is opened
	if workerr, ok := <-w.workerQueue; ok {
		select {
		case workerr <- worker.Task{Work: work, Payload: payload, ResultChan: resultChan}:
			return resultChan, nil
		case <-w.ctx.Done():
			return nil, ErrWorkerPoolClosed2
		}
	}

	return nil, ErrWorkerPoolClosed1
}

func (w *WorkerPool) EnqueueWithTimeout(work worker.Work, payload int, timeout time.Duration) (<-chan int, error) {
	resultChan := make(chan int, 1)

	// Check If Worker Pool is opened
	if workerx, ok := <-w.workerQueue; ok {
		select {
		case workerx <- worker.Task{Work: work, Payload: payload, ResultChan: resultChan}:
			return resultChan, nil
		case <-w.ctx.Done():
			return nil, ErrWorkerPoolClosed2
		case <-time.After(timeout):
			return nil, ErrWorkTimeout
		}
	}
	return nil, ErrWorkerPoolClosed1
}
