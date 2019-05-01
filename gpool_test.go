package gpool_test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/sherifabdlnaby/gpool"
)

// -------------- Testing --------------

func TestPool_Start(t *testing.T) {
	// Test sizes for  < 0, 0 and > 0 size.
	for size := -1; size <= 2; size++ {
		t.Run(fmt.Sprintf("Size[%d]", size), func(t *testing.T) {
			pool, err := gpool.NewPool(size)

			if size < 1 && err == nil {
				t.Errorf("pool construction succeeded with invalid size")
			}

			if size < 1 && err != nil {
				if err != gpool.ErrPoolInvalidSize {
					t.Errorf("pool construction failed but returned incorrect error")
				}
				return
			}

			if size >= 1 && err != nil {
				t.Errorf("pool construction failed, error: %s", err)
			}

			/// Send Work before Worker Start
			Err := pool.Enqueue(context.TODO(), func() {})

			if Err == nil {
				t.Error("Pool Enqueued a Job before pool starts.")
			}

			if Err != gpool.ErrPoolClosed {
				t.Error("Pool Sent an incorrect error type")
			}

			/// Send Work before Worker Start with wait
			Err = pool.EnqueueAndWait(context.TODO(), func() {})

			if Err == nil {
				t.Error("Pool Enqueued a Job before pool starts.")
			}

			if Err != gpool.ErrPoolClosed {
				t.Error("Pool Sent an incorrect error type")
			}

			/// Start Pool
			pool.Start()

			// Test subsequent Calls to Start too
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

	pool, _ := gpool.NewPool(10)

	/// Start Worker
	pool.Start()
	pool.Stop()

	// test subsequent calls to Stop()
	pool.Stop()

	x := make(chan int)
	Err := pool.Enqueue(context.TODO(), func() {
		x <- 123
	})

	if Err == nil {
		t.Errorf("Accepted Job after Stopping the pool")
	}
	if Err != gpool.ErrPoolClosed {
		t.Errorf("Returned Incorrect Error after sending job to stopped pool")
	}
}

func TestPool_Restart(t *testing.T) {

	pool, _ := gpool.NewPool(1)
	/// Start Worker
	pool.Start()

	/// Restarting the Pool
	pool.Stop()

	/// Send Work to pool_closed Pool
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
}

func TestPool_Enqueue(t *testing.T) {

	pool, _ := gpool.NewPool(2)
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
}

func TestPool_EnqueueAndWait(t *testing.T) {

	pool, _ := gpool.NewPool(2)
	// Start Worker
	pool.Start()

	// Enqueue a Job
	x := make(chan int)

	// Enqueue
	done := make(chan struct{})

	go func() {
		Err := pool.EnqueueAndWait(context.TODO(), func() {
			x <- 123
		})

		if Err != nil {
			t.Errorf("Error returned in a started and free pool. Error: %s", Err.Error())
		}

		done <- struct{}{}
	}()

	select {
	case <-done:
		t.Errorf("Receviced Done BEFORE job has returned!")
	case result := <-x:
		if result != 123 {
			t.Errorf("Wrong Result by Job")
		}
	}
}

func TestPool_EnqueueBlocking(t *testing.T) {

	pool, _ := gpool.NewPool(2)

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
}

func TestPool_EnqueueAndWaitBlocking(t *testing.T) {

	pool, _ := gpool.NewPool(1)

	// Create Context
	ctx := context.TODO()

	// Start Worker
	pool.Start()

	/// TEST BLOCKING WHEN POOL IS FULL
	fill := make(chan int)
	Err1 := pool.Enqueue(ctx, func() { fill <- 123 })

	// Enqueue a job with a canceled ctx
	canceledCtx, cancel := context.WithCancel(context.TODO())
	cancel()

	a := make(chan int)
	Err1 = pool.EnqueueAndWait(canceledCtx, func() { a <- 123 })

	if Err1 == nil {
		t.Error("Didn't Return Error in a waiting & canceled job")
	}

}

func TestPool_TryEnqueue(t *testing.T) {

	pool, _ := gpool.NewPool(2)
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
}

func TestSemaphorePool_TryEnqueueAndWait(t *testing.T) {

	pool, _ := gpool.NewPool(2)
	x := make(chan int, 1)

	/// Start Worker
	pool.Start()

	// Enqueue
	done := make(chan struct{})

	go func() {
		success := pool.TryEnqueueAndWait(func() {
			x <- 123
		})

		if !success {
			t.Errorf("False returned in a started and free pool.")
		}

		done <- struct{}{}
	}()

	select {
	case <-done:
		t.Errorf("Receviced Done BEFORE job has returned!")
	case result := <-x:
		if result != 123 {
			t.Errorf("Wrong Result by Job")
		}
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

	success3 := pool.TryEnqueueAndWait(func() { c <- 123 })

	if success3 == true {
		t.Errorf("TryEnqueue success on a FILLED queue")
	}

	<-a
	<-b
}

func TestPool_GetSize(t *testing.T) {

	size := 10
	pool, _ := gpool.NewPool(size)
	pool.Start()
	if pool.GetSize() != size {
		t.Errorf("GetSize() returned incorrect size")
	}

	size = 5
	_ = pool.Resize(size)
	if pool.GetSize() != size {
		t.Errorf("GetSize() returned incorrect size")
	}

	size = 15
	_ = pool.Resize(size)
	if pool.GetSize() != size {
		t.Errorf("GetSize() returned incorrect size")
	}
}

func TestPool_Resize(t *testing.T) {
	size := 10
	pool, _ := gpool.NewPool(size)

	// resize to new size
	size = 0
	err := pool.Resize(size)
	if err == nil {
		t.Errorf("Resize to invalid size didn't return error!")
	}
	if err != gpool.ErrPoolInvalidSize {
		t.Errorf("Resize to invalid size returned wrong error!")
	}

	size = 15
	err = pool.Resize(size)
	if err != nil {
		t.Errorf("Resize failed error: %v", err.Error())
	}

	pool.Start()

	if pool.GetSize() != size {
		t.Errorf("resize didn't return correct size")
	}
}

func TestPool_PositiveResizeLive(t *testing.T) {
	size := 2
	pool, _ := gpool.NewPool(size)
	pool.Start()

	// Create Context
	ctx := context.TODO()

	/// TEST BLOCKING WHEN POOL IS FULL
	a := make(chan int)
	b := make(chan int)
	c := make(chan int)

	/// SEND 3 JOBS (  TWO TO FILL THE POOL, A ONE TO BE CANCELED BY CTX, AND ONE TO WAIT THE FIRST TWO )

	// Two Jobs
	_ = pool.Enqueue(ctx, func() { a <- 123 })
	_ = pool.Enqueue(ctx, func() { b <- 123 })

	// Send a job that will block
	go func() {
		_ = pool.Enqueue(ctx, func() { c <- 123 })
	}()

	select {
	case <-c:
		t.Error("Job Finished BEFORE jobs that should have been blocking this job.")
	default:
		// job C is blocked, now resize should unblock it.
		_ = pool.Resize(pool.GetSize() + 1)

		select {
		case <-c:
		// Give some time for the job to be picked.
		case <-time.After(500 * time.Millisecond):
			t.Error("Job Blocked after resize.")
		}
	}

	<-a
	<-b
}

func TestPool_NegativeResizeLive(t *testing.T) {

	size := 3
	pool, _ := gpool.NewPool(size)
	pool.Start()

	// Create Context
	ctx := context.TODO()

	/// TEST BLOCKING WHEN POOL IS FULL
	a := make(chan int)
	b := make(chan int)
	c := make(chan int)

	/// SEND 3 JOBS
	// Two Jobs
	_ = pool.Enqueue(ctx, func() { a <- 123 })
	_ = pool.Enqueue(ctx, func() { b <- 123 })

	_ = pool.Resize(pool.GetSize() - 1)

	// Now this should block
	go func() {
		_ = pool.Enqueue(ctx, func() { c <- 123 })
	}()

	select {
	case <-c:
		t.Error("Job Finished BEFORE jobs that should have been blocking this job.")
	default:

	}

	// Get all results.
	<-a
	<-b
	<-c
}

func TestPool_Getters(t *testing.T) {

	size := 2
	pool, _ := gpool.NewPool(size)
	pool.Start()

	if pool.GetSize() != size {
		t.Error("Incorrect size pool.")
	}

	if pool.GetWaiting() != 0 {
		t.Error("Incorrect number of waiting jobs")
	}

	if pool.GetCurrent() != 0 {
		t.Error("Incorrect current of an empty pool.")
	}

	// Create Context
	ctx := context.TODO()

	/// TEST BLOCKING WHEN POOL IS FULL
	a := make(chan int)
	b := make(chan int)
	c := make(chan int)

	/// SEND 3 JOBS (  TWO TO FILL THE POOL, A ONE TO BE CANCELED BY CTX, AND ONE TO WAIT THE FIRST TWO )

	// Two Jobs
	_ = pool.Enqueue(ctx, func() { a <- 123 })
	_ = pool.Enqueue(ctx, func() { b <- 123 })

	// Send a job that will block
	go func() {
		_ = pool.Enqueue(ctx, func() { c <- 123 })
	}()

	// give some time to above go func to run (not clean but can't think of a more deterministic approach for now)
	time.Sleep(50 * time.Millisecond)
	if pool.GetWaiting() != 1 {
		t.Error("Incorrect number of waiting jobs")
	}

	if pool.GetSize() != pool.GetCurrent() {
		t.Error("Size doesn't match Current of a filled pool.")
	}
}

// --------------------------------------

// ------------ Benchmarking ------------

func BenchmarkThroughput(b *testing.B) {
	var workersCountValues = []int{10, 100, 1000, 10000}
	for _, workercount := range workersCountValues {
		b.Run(fmt.Sprintf("PoolSize[%d]", workercount), func(b *testing.B) {
			pool, _ := gpool.NewPool(workercount)

			pool.Start()

			b.ResetTimer()
			b.StartTimer()

			for i2 := 0; i2 < b.N; i2++ {
				_ = pool.Enqueue(context.TODO(), func() {
				})
			}

			b.StopTimer()
			pool.Stop()
		})
	}
}

func BenchmarkBulkJobs_UnderLimit(b *testing.B) {
	var workersCountValues = []int{10000}
	var workAmountValues = []int{100, 1000, 10000}

	for _, workercount := range workersCountValues {
		for _, work := range workAmountValues {
			b.Run(fmt.Sprintf("PoolSize[%d]BulkJobs[%d]", workercount, work), func(b *testing.B) {
				pool, _ := gpool.NewPool(workercount)
				pool.Start()
				b.ResetTimer()

				for i2 := 0; i2 < b.N; i2++ {
					wg := sync.WaitGroup{}
					wg.Add(work)
					for i3 := 0; i3 < work; i3++ {
						go func() {
							_ = pool.Enqueue(context.TODO(), func() {})
							wg.Done()
						}()
					}
					wg.Wait()
				}

				b.StopTimer()
				pool.Stop()
			})
		}
	}
}

func BenchmarkBulkJobs_OverLimit(b *testing.B) {
	var workersCountValues = []int{100, 1000}
	var workAmountValues = []int{1000, 10000}

	for _, workercount := range workersCountValues {
		for _, work := range workAmountValues {
			b.Run(fmt.Sprintf("PoolSize[%d]BulkJobs[%d]", workercount, work), func(b *testing.B) {
				pool, _ := gpool.NewPool(workercount)
				pool.Start()
				b.ResetTimer()

				for i2 := 0; i2 < b.N; i2++ {
					wg := sync.WaitGroup{}
					wg.Add(work)
					for i3 := 0; i3 < work; i3++ {
						go func() {
							_ = pool.Enqueue(context.TODO(), func() {})
							wg.Done()
						}()
					}
					wg.Wait()
				}

				b.StopTimer()
				pool.Stop()
			})
		}
	}
}

// --------------------------------------

// --------------EXAMPLES----------------

// Example 1 - Simple Job Enqueue
func Example_one() {
	concurrency := 2

	// Create and start pool.
	pool, err := gpool.NewPool(concurrency)

	if err != nil {
		panic(err)
	}

	pool.Start()

	defer pool.Stop()

	// Create JOB
	resultChan1 := make(chan int)
	ctx := context.Background()
	job := func() {
		time.Sleep(2000 * time.Millisecond)
		resultChan1 <- 1337
	}

	// Enqueue Job
	err1 := pool.Enqueue(ctx, job)

	if err1 != nil {
		log.Printf("Job was not enqueued. Error: [%s]", err1.Error())
		return
	}

	log.Printf("Job Enqueued and started processing")

	log.Printf("Job Done, Received: %v", <-resultChan1)
}

// Example 2 - Enqueue A Job with Timeout
func Example_two() {
	concurrency := 2

	// Create and start pool.
	pool, err := gpool.NewPool(concurrency)

	if err != nil {
		panic(err)
	}

	pool.Start()

	defer pool.Stop()

	// Create JOB
	resultChan := make(chan int)
	ctx := context.Background()
	job := func() {
		resultChan <- 1337
	}

	// Enqueue 2 Jobs to fill pool (Will not finish unless we pull result from resultChan)
	_ = pool.Enqueue(ctx, job)
	_ = pool.Enqueue(ctx, job)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()

	// Will block for 1 second only because of Timeout
	err1 := pool.Enqueue(ctxWithTimeout, job)

	if err1 != nil {
		log.Printf("Job was not enqueued. Error: [%s]", err1.Error())
	}

	log.Printf("Job 1 Done, Received: %v", <-resultChan)
	log.Printf("Job 2 Done, Received: %v", <-resultChan)
}

// Example 3 - Enqueue 10 Jobs and Stop pool mid-processing.
func Example_three() {
	// Create and start pool.
	pool, err := gpool.NewPool(2)

	if err != nil {
		panic(err)
	}

	log.Println("Starting Pool...")

	pool.Start()
	defer pool.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for i := 0; i < 10; i++ {

			// Small Interval for more readable output
			time.Sleep(500 * time.Millisecond)

			go func(i int) {
				x := make(chan int, 1)

				log.Printf("Job [%v] Enqueueing", i)

				err := pool.Enqueue(ctx, func() {
					time.Sleep(2000 * time.Millisecond)
					x <- i
				})

				if err != nil {
					log.Printf("Job [%v] was not enqueued. [%s]", i, err.Error())
					return
				}

				log.Printf("Job [%v] Enqueue-ed ", i)

				log.Printf("Job [%v] Receieved, Result: [%v]", i, <-x)
			}(i)
		}
	}()

	// Uncomment to demonstrate ctx cancel of jobs.
	//time.Sleep(100 * time.Millisecond)
	//cancel()

	time.Sleep(5000 * time.Millisecond)

	fmt.Println("Stopping...")

	pool.Stop()

	fmt.Println("Stopped")

	fmt.Println("Sleeping for couple of seconds so canceled job have a chance to print out their status")

	time.Sleep(4000 * time.Millisecond)
}
