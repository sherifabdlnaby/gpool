# gpool - a generic context-aware resizable goroutines pool to bound concurrency.


[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#goroutines)
<a>
  <img src="https://img.shields.io/github/v/tag/sherifabdlnaby/gpool?label=release&amp;sort=semver">
</a>
[![](https://godoc.org/github.com/sherifabdlnaby/gpool?status.svg)](http://godoc.org/github.com/sherifabdlnaby/gpool)
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

Easily manages a resizeable pool of context aware goroutines to bound concurrency, A **`Job`** is **Enqueued** to the pool and only **`N`** jobs can be processing concurrently at the same time.

- A Job is simply a `func(){}`,  when you `Enqueue(..)` a job, the enqueue call will return *ONCE* the job has **started** processing. Otherwise if the pool is full it will **block** until:
  1. pool has room for the job.
  2. job's `context` is canceled.
  3. the pool is stopped.

- The Pool can be re-sized using `Resize()` that will resize the pool in a concurrent safe-way. `Resize` can enlarge the pool and any blocked enqueue will unblock after pool is resized, in case of shrinking the pool `resize` will not affect any already processing/waiting jobs.

- Enqueuing a Job will return error `nil` once a job starts, `ErrPoolClosed` if the pool is closed, or **the context's error if the job's context is canceled while blocking waiting for the pool.**

- Stopping the Pool using `pool.Stop()` will **wait** for all processing jobs to finish before returning, it will also unblock any **blocked** job enqueues (enqueues will return ErrPoolClosed).

- `Start`, `Stop`, and `Resize(N)` are all concurrent safe and can be called from multiple goroutines, subsequent calls of Start or Stop has no effect unless called interchangeably.

further documentation at : [![](https://godoc.org/github.com/sherifabdlnaby/gpool?status.svg)](http://godoc.org/github.com/sherifabdlnaby/gpool)

------------------------------------------------------

## Usage

- Create new pool
    ```
    pool, err := gpool.NewPool(concurrency)
    ```
- Enqueue a job
    ```
        job := func() {
            time.Sleep(2000 * time.Millisecond)
            fmt.Println("did some work")
        }

        // Enqueue Job
        err := pool.Enqueue(ctx, job)
    ```
    A call to `pool.Enqueue()` will return `nil` if `job` started processing, blocks if the pool is full, `ctx.Err()` if context was canceled while waiting/blocking, or finally `ErrPoolClosed` if the pool stopped or was never started.
- Resize the pool
    ```
    err = pool.Resize(size)
    ```
    Will live change the size of the pool, If new size is larger, waiting job enqueues from another goroutines will be unblocked to fit the new size, and if new size is smaller, any new enqueues will block until the current size of the pool is less than the new one.
- Stop the pool
    ```
    pool.Stop()
    ```
    - ALL Blocked/Waiting jobs will return immediately.

    - Stop() WILL Block until all running jobs is done.

- Different types of "Enqueues"
    - `Enqueue(ctx, job)`          returns ONCE the job has started executing (not after job finishes/return)

    - `EnqueueAndWait(ctx, job)`   returns ONCE the job has started **and** finished executing.

    - `TryEnqueue(job)`            will not block if the pool is full, returns `true` ONCE the job has started executing and `false` if pool is full.

    - `TryEnqueueAndWait(job)`     will not block if the pool is full, returns `true` ONCE the job has started **and** finished executing. and `false` if pool is full.


------------------------------------------------------

## Benchmarks

``` bash
$ go test -bench=. -cpu=2 -benchmem
```

```
go test -bench=. -cpu=2 -benchmem
goos: darwin
goarch: amd64
pkg: github.com/sherifabdlnaby/gpool
BenchmarkThroughput/PoolSize[2]-2                 853724               730 ns/op             125 B/op          2 allocs/op
BenchmarkThroughput/PoolSize[10]-2               3647638               329 ns/op              10 B/op          0 allocs/op
BenchmarkThroughput/PoolSize[100]-2              4869789               248 ns/op               0 B/op          0 allocs/op
BenchmarkThroughput/PoolSize[1000]-2             4320566               280 ns/op               0 B/op          0 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[2]BulkJobs[2]-2             530328              2146 ns/op             275 B/op          5 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[2]BulkJobs[100]-2            10000            112478 ns/op           12737 B/op        239 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[2]BulkJobs[1000]-2            1069           1109384 ns/op          126943 B/op       2380 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[10]BulkJobs[2]-2           1417808               844 ns/op              58 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[10]BulkJobs[100]-2           30556             39187 ns/op            1764 B/op         33 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[10]BulkJobs[1000]-2           3012            385652 ns/op           15737 B/op        295 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[100]BulkJobs[2]-2          1986571               609 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[100]BulkJobs[100]-2          41235             29004 ns/op              37 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[100]BulkJobs[1000]-2          4080            290791 ns/op             188 B/op          4 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[1000]BulkJobs[2]-2         1963564               605 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[1000]BulkJobs[100]-2         42000             28442 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[1000]BulkJobs[1000]-2         4333            284865 ns/op              17 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[10000]BulkJobs[2]-2        1963168               611 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[10000]BulkJobs[100]-2        42238             28419 ns/op              16 B/op          1 allocs/op
BenchmarkBulkJobs_UnderLimit/PoolSize[10000]BulkJobs[1000]-2        4171            283981 ns/op              29 B/op          1 allocs/op
```


**BenchmarkOneThroughput/PoolSize[S]**                    = Enqueue Async Jobs ( Will not wait for result ) in a Pool of size = `S`

**BenchmarkBulkJobs/PoolSize[S]BulkJobs[J]**              = Enqueue `J` Jobs In Pool of size `S` at a time where `J` < `S`


------------------------------------------------------

## Examples

### Example 1 - Simple Job Enqueue
```go
func main() {
    concurrency := 2

    // Create and start pool.
    pool, _ := gpool.NewPool(concurrency)
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
  pool, _ := gpool.NewPool(concurrency)
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
const concurrency = 2

func main() {
  pool, _ := gpool.NewPool(concurrency)
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

# License
[MIT License](https://raw.githubusercontent.com/sherifabdlnaby/gpool/blob/master/LICENSE)
Copyright (c) 2019 Sherif Abdel-Naby

# Contribution

PR(s) are Open and Welcomed.