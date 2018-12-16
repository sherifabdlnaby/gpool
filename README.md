# gpool - A Generic bounded concurrency goroutine pool 

[![](https://godoc.org/github.com/SherifAbdlNaby/gpool?status.svg)](http://godoc.org/github.com/SherifAbdlNaby/gpool)
[![Go Report Card](https://goreportcard.com/badge/github.com/SherifAbdlNaby/gpool)](https://goreportcard.com/report/github.com/SherifAbdlNaby/gpool)
[![Build Status](https://travis-ci.org/Sherifabdlnaby/gpool.svg?branch=func)](https://travis-ci.org/Sherifabdlnaby/gpool)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/SherifAbdlNaby/gpool/blob/master/LICENSE)


Easily manages a pool of context aware goroutines to bound concurrency, A Job is Enqueued to the pool and only `N` jobs can be processeing concurrently.

When you `Enqueue` a job it will return ONCE the jobs **starts** processing otherwise if the pool is full it will **block** until: **(1)** pool has room for the job, **(2)** job's `context` is canceled, or **(3)** the pool is stopped.

Stopping the Pool using `pool.Stop()` will unblock any **blocked** enqueues and **wait** All active jobs to finish before returining.

Enqueuing a Job will return error `nil` once a job starts, `gpool.ErrPoolClosed` if the pool is closed, or `gpool.ErrTimeoutJob` if the job's context is cancled whil blocking waiting for the job.

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
Semaphore is substainally better.


```
goos: darwin
goarch: amd64
pkg: gpool
BenchmarkOneJob/[Workerpool]W[10]-2              1000000              1721 ns/op             128 B/op          2 allocs/op
BenchmarkOneJob/[Workerpool]W[100]-2             1000000              1722 ns/op             128 B/op          2 allocs/op
BenchmarkOneJob/[Workerpool]W[1000]-2            1000000              1729 ns/op             128 B/op          2 allocs/op
BenchmarkOneJob/[Workerpool]W[10000]-2            500000              2560 ns/op             128 B/op          2 allocs/op
BenchmarkOneJob/[SemaphorePool]W[10]-2           1000000              1089 ns/op             128 B/op          2 allocs/op
BenchmarkOneJob/[SemaphorePool]W[100]-2          1000000              1092 ns/op             128 B/op          2 allocs/op
BenchmarkOneJob/[SemaphorePool]W[1000]-2         1000000              1087 ns/op             128 B/op          2 allocs/op
BenchmarkOneJob/[SemaphorePool]W[10000]-2        1000000              1091 ns/op             128 B/op          2 allocs/op
BenchmarkBulkJobs/[Workerpool]W[10]J[1000]-2                2000           1054296 ns/op          128044 B/op       2001 allocs/op
BenchmarkBulkJobs/[Semaphore_]W[10]J[1000]-2                2000            602012 ns/op          128047 B/op       2001 allocs/op
BenchmarkBulkJobs/[Workerpool]W[10]J[10000]-2                100          12345006 ns/op         1280016 B/op      20001 allocs/op
BenchmarkBulkJobs/[Semaphore_]W[10]J[10000]-2                200           7715572 ns/op         1290186 B/op      20113 allocs/op
BenchmarkBulkJobs/[Workerpool]W[10]J[100000]-2                10         128414983 ns/op        12800035 B/op     200001 allocs/op
BenchmarkBulkJobs/[Semaphore_]W[10]J[100000]-2                20         101069406 ns/op        12975425 B/op     200430 allocs/op
BenchmarkBulkJobs/[Workerpool]W[10000]J[1000]-2             1000           1417746 ns/op          128019 B/op       2001 allocs/op
BenchmarkBulkJobs/[Semaphore_]W[10000]J[1000]-2             3000            618805 ns/op          128016 B/op       2001 allocs/op
BenchmarkBulkJobs/[Workerpool]W[10000]J[10000]-2             100          16056593 ns/op         1280204 B/op      20002 allocs/op
BenchmarkBulkJobs/[Semaphore_]W[10000]J[10000]-2             200           7314480 ns/op         1280036 B/op      20001 allocs/op
BenchmarkBulkJobs/[Workerpool]W[10000]J[100000]-2             10         177255711 ns/op        13081756 B/op     200755 allocs/op
BenchmarkBulkJobs/[Semaphore_]W[10000]J[100000]-2             20         101682123 ns/op        12820804 B/op     200217 allocs/op
PASS
ok      gpool   36.814s
```


**W**[*n*] <- How many Worker/Concurrency working at the same time.

**J**[*n*] <- Number of Jobs
