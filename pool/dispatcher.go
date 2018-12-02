package pool

import (
	"fmt"
)

type WorkerPool struct {
	Work        chan Work
	End         chan struct{}
	WorkerQueue chan chan Work
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
