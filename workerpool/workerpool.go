package workerpool

import (
	"context"
	"errors"
	"sync"
)

type Worker struct {
	ID      int
	Worker  chan chan func()
	Receive chan func()
}

type WorkerPool struct {
	WorkerCount int
	workerQueue chan chan func()
	workers     []Worker
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

var (
	//DIFFERENT NAME FOR DEBUGGING ONLY //TODO only1
	ErrWorkerPoolNotStarted = errors.New("pool is closed")
	ErrWorkerPoolClosed1    = errors.New("pool is closed ( By checking  w.ctx.Done() )")
	ErrWorkerPoolClosed12   = errors.New("pool is closed ( By checking w.ctx.Done() [OUTER])")
	ErrWorkerPoolClosed2    = errors.New("pool is closed ( By checking <-ctx.Done() )")
	ErrWorkTimeout          = errors.New("timeout")
)

func NewWorkerPool(workerCount int) *WorkerPool {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	newWorkerPool := WorkerPool{
		WorkerCount: workerCount,
		ctx:         ctx,
		cancel:      cancel,
		workerQueue: make(chan chan func(), workerCount),
		workers:     make([]Worker, workerCount),
	}
	return &newWorkerPool
}

func (w *WorkerPool) Start() {
	// Spin Up Workers
	for i := 0; i < w.WorkerCount; i++ {
		//log.Println(fmt.Sprintf("Starting worker [%d]...", i))

		workerx := Worker{
			ID:      i,
			Receive: make(chan func()),
			Worker:  w.workerQueue,
		}

		// Start worker and start consuming
		workerx.Start(w.ctx, &w.wg)

		// Store workers
		w.workers = append(w.workers, workerx)
	}
}

func (w *WorkerPool) Stop() {
	w.cancel()
	w.wg.Wait()
	close(w.workerQueue)
}

func (w *WorkerPool) Enqueue(ctx context.Context, f func()) error {

	// Check If Worker Pool is opened
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-w.ctx.Done():
		return w.ctx.Err()
	case workerr := <-w.workerQueue:
		select {
		case workerr <- f:
			return nil
		case <-w.ctx.Done():
			return w.ctx.Err()
		}
	}
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
				task()
			case <-ctx.Done():
				//log.Println(fmt.Sprintf("Worker [%d] has stopped.", w.ID))
				return
			}
		}
	}()
}
