package pool

import (
	"fmt"
)

type WorkerPool struct {
	Work        chan Work
	End         chan struct{}
	WorkerQueue chan chan Work
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

func (wp *WorkerPool) StartWorkerPool(workerCount int) {
	var i int
	var workers []Worker

	wp.Work = make(chan Work)    // channel to receive work
	wp.End = make(chan struct{}) // channel to spin down workers
	wp.WorkerQueue = make(chan chan Work, workerCount)

	// Spin Up Workers
	for ; i < workerCount; i++ {
		fmt.Println("starting worker: ", i)
		worker := Worker{
			ID:      i,
			Receive: make(chan Work),
			Worker:  wp.WorkerQueue,
			End:     wp.End,
		}
		worker.Start()
		workers = append(workers, worker) // store worker
	}

	// Start Pool
	go func() {
		for {
			select {
			case work := <-wp.Work:
				workerChan := <-wp.WorkerQueue // wait for available worker receive channel
				workerChan <- work             // dispatch work to worker
			}
		}
	}()

}

func (wp *WorkerPool) StopWorkers() {
	close(wp.End)
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
