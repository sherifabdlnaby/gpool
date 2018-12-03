package workerpool

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type worker struct {
	ID      int
	Worker  chan chan task
	Receive chan task
}

// start worker
func (w *worker) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			w.Worker <- w.Receive
			select {
			case task := <-w.Receive:
				// do work
				task.resultChan <- task.work.Run(task.payload)
				close(task.resultChan)
			case <-ctx.Done():
				log.Println(fmt.Sprintf("worker [%d] has stopped.", w.ID))
				return
			}
		}
	}()
}
