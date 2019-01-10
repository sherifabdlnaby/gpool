# gpool - A Generic bounded concurrency goroutine pool 

[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#goroutines)  [![](https://godoc.org/github.com/SherifAbdlNaby/gpool?status.svg)](http://godoc.org/github.com/SherifAbdlNaby/gpool)
[![Go Report Card](https://goreportcard.com/badge/github.com/sherifabdlnaby/gpool)](https://goreportcard.com/report/github.com/sherifabdlnaby/gpool)
[![Build Status](https://travis-ci.org/sherifabdlnaby/gpool.svg?branch=func)](https://travis-ci.org/sherifabdlnaby/gpool)
[![codecov](https://codecov.io/gh/sherifabdlnaby/gpool/branch/func/graph/badge.svg)](https://codecov.io/gh/sherifabdlnaby/gpool)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/SherifAbdlNaby/gpool/blob/master/LICENSE)

## Installation
``` bash
$ go get github.com/sherifabdlnaby/gpool
```
``` go
import "github.com/sherifabdlnaby/gpool"
```

## Introduction

Easily manages a resizeable pool of context aware goroutines to bound concurrency, A **`Job`** is **Enqueued** to the pool and only **`N`** jobs can be processing concurrently.

- A Job is simply a `func(){}` When you `Enqueue(ctx, func(){})` a job the call will return *ONCE* the job has **started** processing. Otherwise if the pool is full it will **block** until:
  1. pool has room for the job.
  2. job's `context` is canceled.
  3. the pool is stopped.


- A Pool is either `closed` or `started`, the Pool will not accept any job unless `pool.Start()` is called.

  Stopping the Pool using `pool.Stop()` it will **wait** for all processing jobs to finish before returning, it will also unblock any **blocked** job enqueues (enqueues will return ErrPoolClosed).

- The Pool can be re-sized using `Resize()` that will resize the pool in a concurrent safe-way. `Resize` can enlarge the pool and any blocked enqueue will unblock after pool is resized, in case of shrinking the pool `resize` will not affect any already processing job.

- Enqueuing a Job will return error `nil` once a job starts, `ErrPoolClosed` if the pool is closed, or `ErrJobCanceled` if the job's context is canceled while blocking waiting for the pool.

- `Start`, `Stop`, and `Resize(N)` are all concurrent safe and can be called from multiple goroutines, subsequent calls of Start or Stop has no effect unless called interchangeably.

#### Two Implementation
gPool has two implementation for the same Pool Interface{} and both has the same exact behavior, Implementation 1: uses workerpool pattern and 2: uses Semaphore.
According to benchmarks below Semaphore has significantly less overhead than workerpool.

further documentation at : [![](https://godoc.org/github.com/SherifAbdlNaby/gpool?status.svg)](http://godoc.org/github.com/SherifAbdlNaby/gpool)

----------


## Examples

### Example 1 - Simple Job Enqueue
```go
func main() {
  concurrency := 2

  // Create and start pool.
  var pool gpool.Pool = gpool.NewSemaphorePool(concurrency)
  err := pool.Start()

  if err != nil {
    panic(err)
  }

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
```
----------

### Example 2 - Enqueue A Job with Timeout
```go

func main() {
  concurrency := 2

  // Create and start pool.
  var pool gpool.Pool = gpool.NewSemaphorePool(concurrency)
  err := pool.Start()

  if err != nil {
    panic(err)
  }

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


  ctxWithTimeout, _ := context.WithTimeout(ctx, 1000 * time.Millisecond)

  // Will block for 1 second only because of Timeout
  err1 := pool.Enqueue(ctxWithTimeout, job)

  if err1 != nil {
    log.Printf("Job was not enqueued. Error: [%s]", err1.Error())
  }

  log.Printf("Job 1 Done, Received: %v", <-resultChan)
  log.Printf("Job 2 Done, Received: %v", <-resultChan)
}
```

----------

### Example 3
``` go
// size Workers / Concurrent jobs of the Pool
const size = 2

func main() {
  var pool gpool.Pool
  pool = gpool.NewSemaphorePool(size)
  log.Println("Starting Pool...")
  err := pool.Start()

  if err != nil {
    panic(err)
  }
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
```
#### Output
```
2019/01/08 20:15:38 Starting Pool...
2019/01/08 20:15:39 Job [0] Enqueueing
2019/01/08 20:15:39 Job [0] Enqueue-ed
2019/01/08 20:15:39 Job [1] Enqueueing
2019/01/08 20:15:39 Job [1] Enqueue-ed
2019/01/08 20:15:40 Job [2] Enqueueing
2019/01/08 20:15:40 Job [3] Enqueueing
2019/01/08 20:15:41 Job [0] Receieved, Result: [0]
2019/01/08 20:15:41 Job [2] Enqueue-ed
2019/01/08 20:15:41 Job [4] Enqueueing
2019/01/08 20:15:41 Job [3] Enqueue-ed
2019/01/08 20:15:41 Job [1] Receieved, Result: [1]
2019/01/08 20:15:41 Job [5] Enqueueing
2019/01/08 20:15:42 Job [6] Enqueueing
2019/01/08 20:15:42 Job [7] Enqueueing
2019/01/08 20:15:43 Job [4] Enqueue-ed
2019/01/08 20:15:43 Job [2] Receieved, Result: [2]
2019/01/08 20:15:43 Job [8] Enqueueing
Stopping...
2019/01/08 20:15:43 Job [7] was not enqueued. [pool is closed]
2019/01/08 20:15:43 Job [5] was not enqueued. [pool is closed]
2019/01/08 20:15:43 Job [6] was not enqueued. [pool is closed]
2019/01/08 20:15:43 Job [3] Receieved, Result: [3]
2019/01/08 20:15:43 Job [8] was not enqueued. [pool is closed]
2019/01/08 20:15:43 Job [9] Enqueueing
Stopped
2019/01/08 20:15:45 Job [4] Receieved, Result: [4]
Sleeping for couple of seconds so canceled job have a chance to print out their status
2019/01/08 20:15:45 Job [9] was not enqueued. [pool is closed]

Process finished with exit code 0
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
