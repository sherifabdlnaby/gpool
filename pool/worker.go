package pool

import (
	"fmt"
	"log"
	"pipeline/work"
	"sync"
)

type Work struct {
	ID     int
	Job    string
	Result chan string
}

type Worker struct {
	ID      int
	Worker  chan chan Work
	Receive chan Work
	End     chan struct{}
	Wg      *sync.WaitGroup
}

// start worker
func (w *Worker) Start() {
	go func() {
		defer w.Wg.Done()
		for {
			w.Worker <- w.Receive
			select {
			case job := <-w.Receive:
				// do work
				job.Result <- work.DoWork(job.Job, job.ID, w.ID)
				close(job.Result)
			case <-w.End:
				log.Println(fmt.Sprintf("Worker [%d] has stopped.", w.ID))
				return
			}
		}
	}()
}

// end worker
func (w *Worker) Stop() {
	log.Printf("worker [%d] is stopping", w.ID)
	w.End <- struct{}{}
}
