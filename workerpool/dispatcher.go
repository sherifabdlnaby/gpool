package workerpool

import (
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
	end         chan struct{}
	workerQueue chan chan task
	workers     []worker
	wg          sync.WaitGroup
}

func NewWorkerPool(workerCount int) *WorkerPool {
	newWorkerPool := WorkerPool{
		WorkerCount: workerCount,
		work:        make(chan task),     // channel to receive work
		end:         make(chan struct{}), // channel to spin down workers
		workerQueue: make(chan chan task, workerCount),
		workers:     make([]worker, workerCount),
	}
	newWorkerPool.wg.Add(workerCount)
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
			End:     w.end,
			Wg:      &w.wg,
		}

		// Start worker and start consuming
		worker.Start()

		// Store workers
		w.workers = append(w.workers, worker)
	}

	// Start Pool
	go func() {
		for {
			select {
			case work := <-w.work:
				workerChan := <-w.workerQueue // wait for available worker receive channel
				workerChan <- work            // dispatch work to worker
			}
		}
	}()
}

func (w *WorkerPool) Stop() {
	close(w.end)
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
