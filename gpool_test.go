package gpool

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

var implementations = []struct {
	name string
	new  func(workerCount int) interface{}
}{
	{name: "Workerpool", new: func(i int) interface{} {
		return NewWorkerPool(i)
	}},
	{name: "SemaphorePool", new: func(i int) interface{} {
		return NewSemaphorePool(i)
	}},
}

func TestPool_Start(t *testing.T) {
	for _, implementation := range implementations {

		pool := implementation.new(1).(Pool)

		t.Run(implementation.name, func(t *testing.T) {
			/// Send Work before Worker Start
			Err := pool.Enqueue(context.TODO(), func() {})

			if Err == nil {
				t.Error("Pool Enqueued a Job before pool starts.")
			}

			if Err != ErrPoolClosed {
				t.Error("Pool Sent an incorrect error type")
			}

			/// Start Worker
			pool.Start()

			// Enqueue a Job
			Err = pool.Enqueue(context.TODO(), func() {})

			if Err != nil {
				t.Errorf("Pool Enqueued Errored after Start. Error: %s", Err.Error())
			}
		})
	}
}

func TestPool_Stop(t *testing.T) {

	for _, implementation := range implementations {

		pool := implementation.new(10).(Pool)

		t.Run(implementation.name, func(t *testing.T) {

			/// Start Worker
			pool.Start()
			pool.Stop()

			x := make(chan int)
			Err := pool.Enqueue(context.TODO(), func() {
				x <- 123
			})

			if Err == nil {
				t.Errorf("Accepted Job after Stopping the pool")
			}
			if Err != ErrPoolClosed {
				t.Errorf("Returned Incorrect Error after sending job to stopped pool")
			}
		})
	}
}

func TestPool_Restart(t *testing.T) {

	for _, implementation := range implementations {

		pool := implementation.new(1).(Pool)

		t.Run(implementation.name, func(t *testing.T) {
			/// Start Worker
			pool.Start()

			/// Restarting the Pool
			pool.Stop()

			/// Send Work to closed Pool
			Err := pool.Enqueue(context.TODO(), func() {})
			if Err == nil {
				t.Error("Enqueued a job on a stopped pool.")
			}

			pool.Start()

			/// Send Work to pool that has been restarted.
			Err = pool.Enqueue(context.TODO(), func() {})
			if Err != nil {
				t.Errorf("Pool Enqueued Errored after restart. Error: %s", Err.Error())
			}
		})
	}
}

func TestPool_Enqueue(t *testing.T) {

	for _, implementation := range implementations {

		pool := implementation.new(2).(Pool)

		t.Run(implementation.name, func(t *testing.T) {
			// Start Worker
			pool.Start()

			// Enqueue a Job
			x := make(chan int, 1)
			Err := pool.Enqueue(context.TODO(), func() {
				x <- 123
			})

			if Err != nil {
				t.Errorf("Error returned in a started and free pool. Error: %s", Err.Error())
			}

			result := <-x

			if result != 123 {
				t.Errorf("Wrong Result by Job")
			}
		})
	}
}

func TestPool_PoolBlocking(t *testing.T) {

	for _, implementation := range implementations {

		pool := implementation.new(2).(Pool)

		t.Run(implementation.name, func(t *testing.T) {

			// Create Context
			ctx := context.TODO()

			// Start Worker
			pool.Start()

			/// TEST BLOCKING WHEN POOL IS FULL
			a := make(chan int)
			b := make(chan int)
			c := make(chan int)
			d := make(chan int)

			/// SEND 4 JOBS (  TWO TO FILL THE POOL, A ONE TO BE CANCELED BY CTX, AND ONE TO WAIT THE FIRST TWO )

			// Two Jobs
			Err1 := pool.Enqueue(ctx, func() { a <- 123 })
			Err2 := pool.Enqueue(ctx, func() { b <- 123 })

			// Enqueue a job with a canceled ctx
			canceledCtx, cancel := context.WithCancel(context.TODO())
			go cancel()
			Err3 := pool.Enqueue(canceledCtx, func() { c <- 123 })

			// Send a waiting Job
			go func() {
				_ = pool.Enqueue(ctx, func() { d <- 123 })
			}()

			if Err1 != nil {
				t.Errorf("Returned Error and it shouldn't #1, Error: %s", Err1.Error())
			}
			if Err2 != nil {
				t.Errorf("Returned Error and it shouldn't #2, Error: %s", Err2.Error())
			}
			if Err3 == nil {
				t.Error("Didn't Return Error in a waiting & canceled job")
			}
			if Err3 != ErrJobCanceled {
				t.Errorf("Canceled Job returned wronge type of Error. Error: %s", Err3.Error())
			}

			// Check that Job C didn't finish before ONE of A & B finish and make room in the pool.
			for i := 0; i < 3; i++ {
				select {
				case <-a:
					if i > 2 {
						t.Error("Job Finished AFTER a job that should have been finished AFTER.")
					}
					i++
				case <-b:
					if i > 2 {
						t.Error("Job Finished AFTER a job that should have been finished AFTER.")
					}
					i++
				case <-c:
					t.Error("Received a result in a job that shouldn't have been run (it was canceled by ctx).")
				case <-d:
					if i < 1 {
						t.Error("Job Finished BEFORE jobs that should have been blocking this job.")
					}
					i++
				}
			}
		})
	}
}

func TestPool_Enqueue0Worker(t *testing.T) {

	for _, implementation := range implementations {

		pool := implementation.new(0).(Pool)

		t.Run(implementation.name, func(t *testing.T) {

			Err := pool.Enqueue(context.TODO(), func() {})

			if Err != ErrPoolClosed {
				t.Errorf("Should Return ErrPoolClosed")
			}
			pool.Start()

			// Enqueue job that will block because 0 workers.
			signal := make(chan error)
			go func() {
				blockedJobErr := pool.Enqueue(context.TODO(), func() {})
				signal <- blockedJobErr
			}()

			// check that the job actually blocked.
			select {
			case <-signal:
				t.Errorf("Job for signal pool of 0 workers didn't block signal call!")
			default:
			}

			// stop the bool
			pool.Stop()

			// check that blocked job unblocked
			select {
			case blockedJobErr := <-signal:
				if blockedJobErr != ErrPoolClosed {
					t.Errorf("Should Return ErrPoolClosed instead returned error: %s", blockedJobErr)
				}
			}

			// Try enqueue job on closed pool again.
			Err = pool.Enqueue(context.TODO(), func() {})
			if Err != ErrPoolClosed {
				t.Errorf("Should Return ErrPoolClosed after Stopping and Not Started Pool of 0 Workers")
			}
		})
	}
}

func TestPool_TryEnqueue(t *testing.T) {
	for _, implementation := range implementations {
		pool := implementation.new(2).(Pool)
		t.Run(implementation.name, func(t *testing.T) {
			x := make(chan int, 1)

			/// Start Worker
			pool.Start()

			success := pool.TryEnqueue(func() {
				x <- 123
			})

			if success != true {
				t.Errorf("TryEnqueue an empty pool failed")
			}

			result := <-x

			if result != 123 {
				t.Errorf("Wrong Result by Job")
			}

			/// TEST BLOCKING
			a := make(chan int)
			b := make(chan int)
			c := make(chan int)

			/// SEND 3 JOBS (  TWO TO FILL THE POOL, AND ONE TO FAIL BECAUSE THE FIRST TWO FILL THE POOL )
			// Two Jobs
			success1 := pool.TryEnqueue(func() { a <- 123 })
			success2 := pool.TryEnqueue(func() { b <- 123 })

			if success1 == false || success2 == false {
				t.Errorf("Failed to TryEnqueue to the MAX pool limit.")
			}

			success3 := pool.TryEnqueue(func() { c <- 123 })

			if success3 == true {
				t.Errorf("TryEnqueue success on a FILLED queue")
			}

			<-a
			<-b
		})
	}
}

func TestPool_TryEnqueue0Worker(t *testing.T) {
	for _, implementation := range implementations {
		pool := implementation.new(0).(Pool)
		t.Run(implementation.name, func(t *testing.T) {
			x := make(chan int, 1)

			/// Start Worker
			pool.Start()

			success := pool.TryEnqueue(func() {
				x <- 123
			})

			if success == true {
				t.Errorf("TryEnqueue success on a WorkerCount=0 queue!")
			}
		})
	}
}

// ------------ Benchmarking ------------

func BenchmarkOneJob(b *testing.B) {
	var workersCountValues = []int{10, 100, 1000, 10000}
	for i := 0; i < 2; i++ {
		for _, workercount := range workersCountValues {
			var workerPool Pool
			var name string
			if i == 0 {
				name = "Workerpool"
			}
			if i == 1 {
				name = "SemaphorePool"
			}

			b.Run(fmt.Sprintf("[%s]W[%d]", name, workercount), func(b *testing.B) {
				if i == 0 {
					workerPool = NewWorkerPool(workercount)
				}
				if i == 1 {
					workerPool = NewSemaphorePool(workercount)
				}

				workerPool.Start()

				b.ResetTimer()

				for i2 := 0; i2 < b.N; i2++ {
					resultChan := make(chan int, 1)
					_ = workerPool.Enqueue(context.TODO(), func() {
						resultChan <- 123
					})
					<-resultChan
				}

				b.StopTimer()
				workerPool.Stop()
			})
		}
	}
}

func BenchmarkBulkJobs(b *testing.B) {
	var workersCountValues = []int{10, 10000}
	var workAmountValues = []int{1000, 10000, 100000}
	for _, workercount := range workersCountValues {
		for _, work := range workAmountValues {
			for i := 0; i < 2; i++ {
				var workerPool Pool
				var name string
				if i == 0 {
					name = "Workerpool"
				}
				if i == 1 {
					name = "Semaphore "
				}

				b.Run(fmt.Sprintf("[%s]W[%d]J[%d]", name, workercount, work), func(b *testing.B) {
					if i == 0 {
						workerPool = NewWorkerPool(workercount)
					}
					if i == 1 {
						workerPool = NewSemaphorePool(workercount)
					}
					workerPool.Start()
					b.ResetTimer()

					for i2 := 0; i2 < b.N; i2++ {
						wg := sync.WaitGroup{}
						wg.Add(work)
						for i3 := 0; i3 < work; i3++ {
							go func() {
								resultChan := make(chan int, 1)
								_ = workerPool.Enqueue(context.TODO(), func() {
									resultChan <- 123
								})
								<-resultChan
								wg.Done()
							}()
						}
						wg.Wait()
					}

					b.StopTimer()
					workerPool.Stop()
				})
			}
		}
	}
}

// --------------------------------------
