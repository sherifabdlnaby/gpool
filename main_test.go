package main

import (
	"context"
	"fmt"
	"pipeline/work"
	"sync"
	"testing"
)

var workersCountValues = []int{10, 10000}
var workAmountValues = []int{10000, 100000}

func BenchmarkCallAndWait(b *testing.B) {
	for i := 0; i < 2; i++ {
		for _, workercount := range workersCountValues {
			var workerPool dpool.Pool
			var name string
			if i == 0 {
				name = "Workerpool"
			}
			if i == 1 {
				name = "Semaphore "
			}

			b.Run(fmt.Sprintf("[%s]W[%d]", name, workercount), func(b *testing.B) {
				if i == 0 {
					workerPool = dpool.NewWorkerPool(workercount)
				}
				if i == 1 {
					workerPool = dpool.NewSemaphorePool(int64(workercount))
				}

				workerPool.Start()

				b.ResetTimer()

				for i2 := 0; i2 < b.N; i2++ {
					resultChan := make(chan int, 1)
					_ = workerPool.Enqueue(context.TODO(), func() {
						resultChan <- sr.Run(123)
					})
					<-resultChan
				}

				b.StopTimer()
				workerPool.Stop()
			})
		}
	}
}

func BenchmarkBulkWait(b *testing.B) {
	for _, workercount := range workersCountValues {
		for _, work := range workAmountValues {
			for i := 0; i < 2; i++ {
				var workerPool dpool.Pool
				var name string
				if i == 0 {
					name = "Workerpool"
				}
				if i == 1 {
					name = "Semaphore "
				}

				b.Run(fmt.Sprintf("[%s]W[%d]J[%d]", name, workercount, work), func(b *testing.B) {
					if i == 0 {
						workerPool = dpool.NewWorkerPool(workercount)
					}
					if i == 1 {
						workerPool = dpool.NewSemaphorePool(int64(workercount))
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
									resultChan <- sr.Run(123)
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

/*func BenchmarkWorkersOverhead(b *testing.B) {

	for i := 0; i < 2; i++ {
		for _, workercount := range workersCountValues {
				var workerPool worker.ConcurrencyBounder
				var name string
				if i == 0 {
					name = "Workerpool"
				}
				if i == 1 {
					name = "Semaphore"
				}

				b.Run(fmt.Sprintf("[%s]W[%d]", name, workercount), func(b *testing.B) {


					b.ResetTimer()

					for i2 := 1; i2 < b.N; i2++ {
						if i == 0 {
							workerPool = workerpool.NewWorkerPool(workercount)
						}
						if i == 1 {
							workerPool = semp.NewSempWorker(int64(workercount))
						}
						workerPool.Start()
						workerPool.Stop()
					}

					b.StopTimer()
				})
		}
	}
}

*/
