package pool

import (
	"fmt"
	"log"
	"pipeline/Payload"
	"sync"
	"time"
)

type WorkerPool struct {
	Work        chan Work
	End         chan struct{}
	WorkerQueue chan chan Work
	workers     []Worker
	wg          sync.WaitGroup
}

func (w *WorkerPool) Init(workerCount int) {

	w.Work = make(chan Work)    // channel to receive work
	w.End = make(chan struct{}) // channel to spin down workers
	w.WorkerQueue = make(chan chan Work, workerCount)
	w.workers = make([]Worker, workerCount)
	w.wg.Add(workerCount)

	// Spin Up Workers
	for i := 0; i < workerCount; i++ {

		fmt.Println("Starting Worker: ", i)

		worker := Worker{
			ID:      i,
			Receive: make(chan Work),
			Worker:  w.WorkerQueue,
			End:     w.End,
			Wg:      &w.wg,
		}

		// Start Worker and start consuming
		worker.Start()

		// store worker
		w.workers = append(w.workers, worker)
	}

}

func (w *WorkerPool) Start() {
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

func (w *WorkerPool) Enqueue(payload Payload.Payload) (<-chan string, error) {
	resultChan := make(chan string, 1)
	w.Work <- Work{Job: payload.Json, ID: payload.ID, Result: resultChan}
	return resultChan, nil
}

func (w *WorkerPool) EnqueueWithTimeout(payload Payload.Payload, timeout time.Duration) (<-chan string, error) {
	resultChan := make(chan string, 1)
	select {
	case w.Work <- Work{Job: payload.Json, ID: payload.ID, Result: resultChan}:
		return resultChan, nil
	case <-time.After(timeout):
		close(resultChan)
		log.Println(fmt.Sprintf("Job [%d] Timedout", payload.ID))
		return resultChan, nil
	}
}
