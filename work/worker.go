package worker

import (
	"context"
	"sync"
)

type Worker struct {
	ID      int
	Worker  chan chan Task
	Receive chan Task
}

type Task struct {
	Work       Work
	Payload    int
	ResultChan chan int
}

type Work interface {
	Run(int) int
}

type ConcurrencyBounder interface {
	Start()
	Stop()
	Enqueue(work Work, payload int) (<-chan int, error)
}

// start Worker
func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			w.Worker <- w.Receive
			select {
			case task := <-w.Receive:
				// do Work
				task.ResultChan <- task.Work.Run(task.Payload)
				close(task.ResultChan)
			case <-ctx.Done():
				//log.Println(fmt.Sprintf("Worker [%d] has stopped.", w.ID))
				return
			}
		}
	}()
}
