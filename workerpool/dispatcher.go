package workerpool

import (
	"context"
	"fmt"
	"log"
	"pipeline/Payload"
	"sync"
	"time"
)

type Work interface {
	Run(Payload.Payload) Payload.Payload
}

type task struct {
	work       Work
	payload    Payload.Payload
	resultChan chan Payload.Payload
}

type WorkerPool struct {
	WorkerCount int
	work        chan task
	workerQueue chan chan task
	workers     []worker
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewWorkerPool(workerCount int) *WorkerPool {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	newWorkerPool := WorkerPool{
		WorkerCount: workerCount,
		work:        make(chan task), // channel to receive work
		ctx:         ctx,
		cancel:      cancel,
		workerQueue: make(chan chan task, workerCount),
		workers:     make([]worker, workerCount),
	}
	return &newWorkerPool
}

func (w *WorkerPool) Start() {
	// Spin Up Workers
	for i := 0; i < w.WorkerCount; i++ {

		log.Println(fmt.Sprintf("Starting worker [%d]...", i))

		worker := worker{
			ID:      i,
			Receive: make(chan task),
			Worker:  w.workerQueue,
		}

		// Start worker and start consuming
		worker.Start(w.ctx, &w.wg)

		// Store workers
		w.workers = append(w.workers, worker)
	}

	// Start Pool
	go func() {
		for {
			<-w.workerQueue <- <-w.work
		}
	}()
}

func (w *WorkerPool) Stop() {
	w.cancel()
	w.wg.Wait()
}

func (w *WorkerPool) Enqueue(work Work, payload Payload.Payload) <-chan Payload.Payload {
	resultChan := make(chan Payload.Payload, 1)
	w.work <- task{work: work, payload: payload, resultChan: resultChan}
	return resultChan
}

func (w *WorkerPool) EnqueueWithTimeout(work Work, payload Payload.Payload, timeout time.Duration) (<-chan Payload.Payload, error) {
	resultChan := make(chan Payload.Payload, 1)
	select {
	case w.work <- task{work: work, payload: payload, resultChan: resultChan}:
		return resultChan, nil
	case <-time.After(timeout):
		close(resultChan)
		log.Println(fmt.Sprintf("Job [%d] Timedout", payload.ID))
		return resultChan, nil
		// TODO Errors
	}
}
