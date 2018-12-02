package workerpool

import (
	"fmt"
	"log"
	"pipeline/Payload"
	"sync"
	"time"
)

type WorkerPool struct {
	WorkerCount int
	Work        chan Task
	End         chan struct{}
	WorkerQueue chan chan Task
	workers     []Worker
	wg          sync.WaitGroup
}

func NewWorkerPool(workerCount int) *WorkerPool {
	newWorkerPool := WorkerPool{
		WorkerCount: workerCount,
		Work:        make(chan Task),     // channel to receive work
		End:         make(chan struct{}), // channel to spin down workers
		WorkerQueue: make(chan chan Task, workerCount),
		workers:     make([]Worker, workerCount),
	}
	newWorkerPool.wg.Add(workerCount)
	return &newWorkerPool
}

func (w *WorkerPool) Start() {
	// Spin Up Workers
	for i := 0; i < w.WorkerCount; i++ {

		log.Println(fmt.Sprintf("Starting Worker [%d]...", i))

		worker := Worker{
			ID:      i,
			Receive: make(chan Task),
			Worker:  w.WorkerQueue,
			End:     w.End,
			Wg:      &w.wg,
		}

		// Start Worker and start consuming
		worker.Start()

		// Store workers
		w.workers = append(w.workers, worker)
	}

	// Start Pool
	go func() {
		for {
			select {
			case work := <-w.Work:
				workerChan := <-w.WorkerQueue // wait for available worker receive channel
				workerChan <- work            // dispatch work to worker
			}
		}
	}()
}

func (w *WorkerPool) Stop() {
	close(w.End)
	w.wg.Wait()
}

func (w *WorkerPool) Enqueue(work Work, payload Payload.Payload) <-chan Payload.Payload {
	resultChan := make(chan Payload.Payload, 1)
	w.Work <- Task{Work: work, Payload: payload, ResultChan: resultChan}
	return resultChan
}

func (w *WorkerPool) EnqueueWithTimeout(work Work, payload Payload.Payload, timeout time.Duration) (<-chan Payload.Payload, error) {
	resultChan := make(chan Payload.Payload, 1)
	select {
	case w.Work <- Task{Work: work, Payload: payload, ResultChan: resultChan}:
		return resultChan, nil
	case <-time.After(timeout):
		close(resultChan)
		log.Println(fmt.Sprintf("Job [%d] Timedout", payload.ID))
		return resultChan, nil
		// TODO Errors
	}
}
