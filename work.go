package main

import "time"

type WorkRequest struct {
	Name  string
	Delay time.Duration
}
