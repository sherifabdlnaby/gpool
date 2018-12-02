package workerpool

import (
	"fmt"
	"log"
	"sync"
)

type worker struct {
	ID      int
	Worker  chan chan task
	Receive chan task
	End     chan struct{}
	Wg      *sync.WaitGroup
}

// start worker
func (w *worker) Start() {
	go func() {
		defer w.Wg.Done()
		for {
			w.Worker <- w.Receive
			select {
			case task := <-w.Receive:
				// do work
				task.resultChan <- task.work.Run(task.payload)
				close(task.resultChan)
			case <-w.End:
				log.Println(fmt.Sprintf("worker [%d] has stopped.", w.ID))
				return
			}
		}
	}()
}

// end worker
func (w *worker) Stop() {
	log.Printf("worker [%d] is stopping", w.ID)
	w.End <- struct{}{}
}
