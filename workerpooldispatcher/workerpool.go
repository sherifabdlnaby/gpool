package workerpooldispatch

import (
	"context"
	"pipeline/work"
	"sync"
	"time"
)

type WorkerPool struct {
	WorkerCount int
	work        chan worker.Task
	workerQueue chan chan worker.Task
	workers     []worker.Worker
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewWorkerPool(workerCount int) *WorkerPool {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	newWorkerPool := WorkerPool{
		WorkerCount: workerCount,
		work:        make(chan worker.Task), // channel to receive work
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

	// Start Pool
	go func() {
		for {
			select {
			case workerChan := <-w.workerQueue:
				work := <-w.work   // wait for available worker receive channel
				workerChan <- work // dispatch work to worker
			case <-w.ctx.Done():
				return
			}
		}
	}()
}

func (w *WorkerPool) Stop() {
	w.cancel()
	w.wg.Wait()
}

func (w *WorkerPool) Enqueue(work worker.Work, payload int) (<-chan int, error) {
	resultChan := make(chan int, 1)
	w.work <- worker.Task{Work: work, Payload: payload, ResultChan: resultChan}
	return resultChan, nil
}

func (w *WorkerPool) EnqueueWithTimeout(work worker.Work, payload int, timeout time.Duration) (<-chan int, error) {
	resultChan := make(chan int, 1)
	select {
	case w.work <- worker.Task{Work: work, Payload: payload, ResultChan: resultChan}:
		return resultChan, nil
	case <-time.After(timeout):
		close(resultChan)
		return resultChan, nil
	}
}
