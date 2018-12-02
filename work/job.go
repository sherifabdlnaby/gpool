package work

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func CreateJobs(amount int) []string {
	var jobs []string

	for i := 0; i < amount; i++ {
		jobs = append(jobs, RandStringRunes(8))
	}
	return jobs
}

// mimics any type of job that can be run concurrently
func DoWork(word string, jobId int, workerId int) string {
	h := fnv.New32a()
	h.Write([]byte(word))
	fmt.Printf("worker [%d] - Job[%d]: created hash [%d] from word [%s]\n", workerId, jobId, h.Sum32(), word)
	time.Sleep(time.Millisecond * 2000)
	return word
}
