package worker

import (
	"context"
	"sync"
)

type Worker struct {
	ID      int
	Worker  chan chan func()
	Receive chan func()
}

type ConcurrencyBounder interface {
	Start()
	Stop()
	Enqueue(f func()) error
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
				task()
			case <-ctx.Done():
				//log.Println(fmt.Sprintf("Worker [%d] has stopped.", w.ID))
				return
			}
		}
	}()
}
