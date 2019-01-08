// gpool is used to easily manages a resizeable pool of context aware goroutines to bound concurrency, A Job is Enqueued to the pool and only N jobs can be processing concurrently.
//
// A Job is simply a func(){} When you Enqueue(ctx, func(){}) a job the call will return ONCE the job has started processing. Otherwise if the pool is full it will block until:
// 1. pool has room for the job.
// 2. job's context is canceled.
// 3. the pool is stopped.
// A Pool is either closed or started, the Pool will not accept any job unless pool.Start() is called.
// Stopping the Pool using pool.Stop() it will wait for all processing jobs to finish before returning, it will also unblock any blocked job enqueues (enqueues will return ErrPoolClosed).
// The Pool can be re-sized using Resize() that will resize the pool in a concurrent safe-way. Resize can enlarge the pool and any blocked enqueue will unblock after pool is resized, in case of shrinking the pool resize will not affect any already processing job.
// Enqueuing a Job will return error nil once a job starts, ErrPoolClosed if the pool is closed, or ErrJobCanceled if the job's context is canceled while blocking waiting for the pool.
// Start, Stop, and Resize(N) is all concurrent safe and can be called from multiple goroutines, subsequent calls of Start or Stop has no effect unless called interchangeably.
package gpool

// Version gpool current interface version.
var Version = "1.0.0"
