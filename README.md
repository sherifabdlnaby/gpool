# gpool - A Generic bounded concurrency goroutine pool 

[![](https://godoc.org/github.com/SherifAbdlNaby/gpool?status.svg)](http://godoc.org/github.com/SherifAbdlNaby/gpool)
[![Go Report Card](https://goreportcard.com/badge/github.com/SherifAbdlNaby/gpool)](https://goreportcard.com/report/github.com/SherifAbdlNaby/gpool)
[![Build Status](https://travis-ci.org/Sherifabdlnaby/gpool.svg?branch=func)](https://travis-ci.org/Sherifabdlnaby/gpool)
[![codecov](https://codecov.io/gh/Sherifabdlnaby/gpool/branch/func/graph/badge.svg)](https://codecov.io/gh/Sherifabdlnaby/gpool)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/SherifAbdlNaby/gpool/blob/master/LICENSE)


Easily manages a resizeable pool of context aware goroutines to bound concurrency, A `Job` is Enqueued to the pool and only `N` jobs can be processing concurrently.

When you `Enqueue` a job it will return ONCE the job **starts** processing otherwise if the pool is full it will **block** until: **(1)** pool has room for the job, **(2)** job's `context` is canceled, or **(3)** the pool is stopped.

Stopping the Pool using `pool.Stop()` will unblock any **blocked** enqueues and **wait** All active jobs to finish before returning.

Enqueuing a Job will return error `nil` once a job starts, `gpool.ErrPoolClosed` if the pool is closed, or `gpool.ErrTimeoutJob` if the job's context is canceled while blocking waiting for the pool.

The Pool can be re-sized using `Resize()` that will resize the pool in a concurrent safe-way. `Resize` can enlarge the pool so that any blocked enqueue will unblock after resize is called, in case of shrinking the pool `resize` will not affect any processing job.

further documentation at : [![](https://godoc.org/github.com/SherifAbdlNaby/gpool?status.svg)](http://godoc.org/github.com/SherifAbdlNaby/gpool)

## Installation
``` bash
$ go get github.com/sherifabdlnaby/gpool
```


## Examples

### Example 1
```go

func main() {
  concurrency := 2
  var pool gpool.Pool = gpool.NewSemaphorePool(concurrency)
  pool.Start()
  defer pool.Stop()

  // Send JOB
  resultChan1 := make(chan int)
  err1 := pool.Enqueue(context.TODO(), func() {
    time.Sleep(2000 * time.Millisecond)
    resultChan1 <- 1337
  })
  if err1 != nil {
    log.Printf("Job was not enqueued. Error: [%s]", err1.Error())
    return
  }

  log.Printf("Job Done, Received: %v", <-resultChan1)
}
```
----------

### Example 2
```go

func main() {
  concurrency := 2
  var pool gpool.Pool = gpool.NewSemaphorePool(concurrency)
  pool.Start()
  defer pool.Stop()

  // Send JOB 1
  resultChan1 := make(chan int)
  err1 := pool.Enqueue(context.TODO(), func() {
    time.Sleep(2000 * time.Millisecond)
    resultChan1 <- 100
  })
  if err1 != nil {
    log.Printf("Job [%v] was not enqueued. [%s]", 1, err1.Error())
    return
  }

  // Send JOB 2
  resultChan2 := make(chan int)
  err2 := pool.Enqueue(context.TODO(), func() {
    time.Sleep(1000 * time.Millisecond)
    resultChan2 <- 200
  })
  if err2 != nil {
    log.Printf("Job [%v] was not enqueued. [%s]", 2, err2.Error())
    return
  }

  // Recieve The Two Jobs which ever come first
  for i := 0; i < 2; i++ {
    select {
    case result := <-resultChan1:
      log.Printf("Job [%v] Done [%v]", 1, result)
    case result := <-resultChan2:
      log.Printf("Job [%v] Done [%v]", 2, result)
    }
  }
}
```

----------

### Example 3
``` go
// WorkerCount Number of Workers / Concurrent jobs of the Pool
const WorkerCount = 2

func main() {
  var workerPool gpool.Pool

  workerPool = gpool.NewSemaphorePool(WorkerCount)

  log.Println("Starting Pool...")

  workerPool.Start()

  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()

  go func() {
    log.Printf("Enqueueing 10 jobs on a seperate goroutine...")
    for i := 0; i < 10; i++ {

      // Small Interval for more readable output
      time.Sleep(500 * time.Millisecond)

      go func(i int) {
        x := make(chan int, 1)

        log.Printf("Job [%v] Enqueueing", i)

        err := workerPool.Enqueue(ctx, func() {
          time.Sleep(2000 * time.Millisecond)
          x <- i
        })

        if err != nil {
          log.Printf("Job [%v] was not enqueued. [%s]", i, err.Error())
          return
        }

        log.Printf("Job [%v] Enqueue-ed ", i)

        log.Printf("Job [%v] Receieved [%v]", i, <-x)
      }(i)
    }
  }()

  // Wait 5 Secs before Stopping
  time.Sleep(5000 * time.Millisecond)

  fmt.Println("Stopping...")

  workerPool.Stop()

  fmt.Println("Stopped")

  fmt.Println("Sleeping for couple of seconds so canceled job have a chance to print out their status")

  time.Sleep(10000 * time.Millisecond)
}
```
#### Output
``` 
2018/12/16 05:37:03 Starting Pool...
2018/12/16 05:37:03 Enqueueing 10 jobs on a seperate goroutine...
2018/12/16 05:37:03 Job [0] Enqueueing
2018/12/16 05:37:03 Job [0] Enqueue-ed 
2018/12/16 05:37:04 Job [1] Enqueueing
2018/12/16 05:37:04 Job [1] Enqueue-ed 
2018/12/16 05:37:04 Job [2] Enqueueing
2018/12/16 05:37:05 Job [3] Enqueueing
2018/12/16 05:37:05 Job [2] Enqueue-ed 
2018/12/16 05:37:05 Job [0] Receieved [0]
2018/12/16 05:37:05 Job [4] Enqueueing
2018/12/16 05:37:06 Job [3] Enqueue-ed 
2018/12/16 05:37:06 Job [1] Receieved [1]
2018/12/16 05:37:06 Job [5] Enqueueing
2018/12/16 05:37:06 Job [6] Enqueueing
2018/12/16 05:37:07 Job [7] Enqueueing
2018/12/16 05:37:07 Job [4] Enqueue-ed 
2018/12/16 05:37:07 Job [2] Receieved [2]
2018/12/16 05:37:07 Job [8] Enqueueing
Stopping...
2018/12/16 05:37:08 Job [3] Receieved [3]
2018/12/16 05:37:08 Job [5] was not enqueued. [pool is closed]
2018/12/16 05:37:08 Job [6] was not enqueued. [pool is closed]
2018/12/16 05:37:08 Job [8] was not enqueued. [pool is closed]
2018/12/16 05:37:08 Job [7] was not enqueued. [pool is closed]
2018/12/16 05:37:08 Job [9] Enqueueing
Stopped
2018/12/16 05:37:09 Job [4] Receieved [4]
Sleeping for couple of seconds so canceled job have a chance to print out their status
2018/12/16 05:37:09 Job [9] was not enqueued. [pool is closed]
```

---------------


## Benchmarks
Benchmarks for the two goroutines pool implementation `Workerpool` & `Semaphore`.
Semaphore is substantially better.

``` bash
$ go test -bench=. -cpu=2 -benchmem
```

**S**[*n*] <- Size of Pool / N Concurrent work at the same time.

**J**[*n*] <- Number of Jobs

**BenchmarkOneThroughput/S[]** = Enqueue Async Jobs ( Will not wait for result ) in a Pool of size = `S`

**BenchmarkOneJobSync/S[]**    = Enqueue One Jobs at a time and wait result (Pool will have at MAX one job running)

**BenchmarkBulkJobs_UnderLimit/S[]J[]**   = Enqueue `J` Jobs In Pool of size `S` at a time where `J` < `S`

**BenchmarkBulkJobs_OverLimit/S[]J[]**    = Enqueue `J` Jobs In Pool of size `S` at a time where `J` > `S`

```
goos: darwin
goarch: amd64
pkg: github.com/sherifabdlnaby/gpool
BenchmarkOneThroughput/[Semaphore]S[10]-2                5000000               258 ns/op              14 B/op          0 allocs/op
BenchmarkOneThroughput/[Semaphore]S[100]-2               5000000               264 ns/op               0 B/op          0 allocs/op
BenchmarkOneThroughput/[Semaphore]S[1000]-2              5000000               279 ns/op               0 B/op          0 allocs/op
BenchmarkOneThroughput/[Semaphore]S[10000]-2             5000000               284 ns/op               0 B/op          0 allocs/op
BenchmarkOneThroughput/[Workerpool]S[10]-2               5000000               367 ns/op               0 B/op          0 allocs/op
BenchmarkOneThroughput/[Workerpool]S[100]-2              5000000               329 ns/op               0 B/op          0 allocs/op
BenchmarkOneThroughput/[Workerpool]S[1000]-2             5000000               325 ns/op               0 B/op          0 allocs/op
BenchmarkOneThroughput/[Workerpool]S[10000]-2            3000000               534 ns/op               0 B/op          0 allocs/op
BenchmarkOneJobSync/[Semaphore]S[10]-2                   2000000               909 ns/op              16 B/op          1 allocs/op
BenchmarkOneJobSync/[Semaphore]S[100]-2                  2000000               911 ns/op              16 B/op          1 allocs/op
BenchmarkOneJobSync/[Semaphore]S[1000]-2                 2000000               914 ns/op              16 B/op          1 allocs/op
BenchmarkOneJobSync/[Semaphore]S[10000]-2                2000000               910 ns/op              16 B/op          1 allocs/op
BenchmarkOneJobSync/[Workerpool]S[10]-2                  1000000              1218 ns/op              16 B/op          1 allocs/op
BenchmarkOneJobSync/[Workerpool]S[100]-2                 1000000              1349 ns/op              16 B/op          1 allocs/op
BenchmarkOneJobSync/[Workerpool]S[1000]-2                1000000              1232 ns/op              16 B/op          1 allocs/op
BenchmarkOneJobSync/[Workerpool]S[10000]-2               1000000              2137 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/[Semaphore]S[10000]J[100]-2                   30000             40378 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/[Semaphore]S[10000]J[1000]-2                   5000            363141 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/[Semaphore]S[10000]J[10000]-2                   300           4042157 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/[Workerpool]S[10000]J[100]-2                  20000             69548 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/[Workerpool]S[10000]J[1000]-2                  2000            714208 ns/op              70 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/[Workerpool]S[10000]J[10000]-2                  200           8868398 ns/op              21 B/op          1 allocs/op
BenchmarkBulkJobs_OverLimit/[Semaphore]S[100]J[1000]-2                      5000            362379 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_OverLimit/[Semaphore]S[100]J[10000]-2                      300           4054593 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_OverLimit/[Semaphore]S[1000]J[1000]-2                     5000            361441 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_OverLimit/[Semaphore]S[1000]J[10000]-2                     300           4057458 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_OverLimit/[Workerpool]S[100]J[1000]-2                     3000            514161 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_OverLimit/[Workerpool]S[100]J[10000]-2                     200           6305356 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_OverLimit/[Workerpool]S[1000]J[1000]-2                    3000            525189 ns/op              20 B/op          1 allocs/op
BenchmarkBulkJobs_OverLimit/[Workerpool]S[1000]J[10000]-2                    200           7027354 ns/op              30 B/op          1 allocs/op
PASS
ok      github.com/sherifabdlnaby/gpool 58.464s
```