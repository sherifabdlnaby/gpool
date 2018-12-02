package workerpool

import (
	"fmt"
	"log"
	"pipeline/Payload"
	"sync"
)

type Work interface {
	Run(Payload.Payload) Payload.Payload
}

type Task struct {
	Work       Work
	Payload    Payload.Payload
	ResultChan chan Payload.Payload
}

type Worker struct {
	ID      int
	Worker  chan chan Task
	Receive chan Task
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
			case task := <-w.Receive:
				// do work
				task.ResultChan <- task.Work.Run(task.Payload)
				close(task.ResultChan)
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
