package gpool

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestWorkerPool_Start(t *testing.T) {

	var implementations = []struct {
		name string
		impl Pool
	}{
		{name: "Workerpool", impl: NewWorkerPool(1)},
		{name: "SemaphorePool", impl: NewSemaphorePool(1)},
	}

	for _, poolImpl := range implementations {

		pool := poolImpl.impl

		t.Run(poolImpl.name, func(t *testing.T) {
			/// Send Work before Worker Start
			Err := pool.Enqueue(context.TODO(), func() {})

			if Err == nil {
				t.Error("Pool Enqueued a Job before pool starts.")
			}

			if Err != ErrPoolClosed {
				t.Error("Pool Sent an incorrect error")
			}

			/// Start Worker
			pool.Start()

			Err = pool.Enqueue(context.TODO(), func() {})

			if Err != nil {
				t.Errorf("Pool Enqueued Errored after Start. Error: %s", Err.Error())
			}

		})
	}
}

func TestWorkerPool_Enqueue(t *testing.T) {

	var implementations = []struct {
		name string
		impl Pool
	}{
		{name: "Workerpool", impl: NewWorkerPool(2)},
		{name: "SemaphorePool", impl: NewSemaphorePool(2)},
	}

	for _, poolImpl := range implementations {

		pool := poolImpl.impl

		t.Run(poolImpl.name, func(t *testing.T) {

			ctx := context.TODO()
			canceledCox, cancel := context.WithCancel(context.TODO())
			cancel()

			/// Start Worker
			pool.Start()

			x := make(chan int, 1)

			Err := pool.Enqueue(context.TODO(), func() {
				x <- 123
			})

			if Err != nil {
				t.Errorf("Shouldn't Return an Error")
			}

			result := <-x

			if result != 123 {
				t.Errorf("Wrong Result by Job")
			}

			/// TEST BLOCKING
			a := make(chan int, 1)
			b := make(chan int, 1)
			c := make(chan int, 1)
			d := make(chan int, 1)

			/// SEND 4 JOBS (  TWO TO FILL THE POOL, A ONE TO BE CANCELED BY CTX, AND ONE TO WAIT THE FIRST TWO )
			// Two Jobs
			Err1 := pool.Enqueue(ctx, func() { time.Sleep(100 * time.Millisecond); a <- 123 })
			Err2 := pool.Enqueue(ctx, func() { time.Sleep(100 * time.Millisecond); b <- 123 })
			// Canceled Job
			Err3 := pool.Enqueue(canceledCox, func() { c <- 123 })
			// Waiting Job
			_ = pool.Enqueue(ctx, func() { time.Sleep(100 * time.Millisecond); d <- 123 })

			if Err1 != nil {
				t.Errorf("Returned Error and it shouldn't #1, Error: %s", Err1.Error())
			}
			if Err2 != nil {
				t.Errorf("Returned Error and it shouldn't #2, Error: %s", Err2.Error())
			}
			if Err3 == nil {
				t.Error("Didn't Return Error in a waiting & canceled job")
			}
			if Err3 != ErrJobTimeout {
				t.Errorf("Canceled Job Timeout returned wronge Error. Error: %s", Err3.Error())
			}

			for i := 0; i < 3; i++ {
				select {
				case <-a:
					if i > 1 {
						t.Error("Job Finished AFTER a job that should have been finished AFTER.")
					}
				case <-b:
					if i > 1 {
						t.Error("Job Finished AFTER a job that should have been finished AFTER.")
					}
				case <-c:
					t.Error("Received a result in a job that shouldn't have been run.")
				case <-d:
					if i < 2 {
						t.Error("Job Finished BEFORE a job that should have been finished.")
					}
				}
			}
		})
	}
}

func BenchmarkOneJob(b *testing.B) {
	var workersCountValues = []int{10, 10000}
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
