# Semaphore

[![](https://godoc.org/github.com/sherifabdlnaby/semaphore?status.svg)](http://godoc.org/github.com/sherifabdlnaby/semaphore)
[![Build Status](https://travis-ci.org/sherifabdlnaby/semaphore.svg?branch=master)](https://travis-ci.org/sherifabdlnaby/semaphore)
[![codecov](https://codecov.io/gh/Sherifabdlnaby/semaphore/branch/resizable-semaphore/graph/badge.svg)](https://codecov.io/gh/Sherifabdlnaby/semaphore)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)

A fork of `golang/x/sync/semaphore` that is also **resizable**.

----

This fork adds the ability to resize the semaphore in a concurrent-safe way.

- Use `semaphore.Resize(N)` to resize the semaphore to a new absolute value `N`
- #### Example
    ```go
    sem := semaphore.NewWeighted(5)     // Semaphore size = 5
    result1 := sem.TryAcquire(10)       // returns false.
    sem.Resize(15)                      // Semaphore size now = 10
    result2 := sem.TryAcquire(10)       // returns true.
    fmt.Println(result1,result2)        // prints false \n true  
    ```
- Added resize functionality has **no impact on the semaphore performance** according to benchmarks.
- Tests for the newly added `Resize()` was also added.