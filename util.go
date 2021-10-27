package gocachelib

import (
	"time"
)

func doEvery(d time.Duration, f func()) *time.Ticker {
	ticker := time.NewTicker(d)
	go func() {
		for range ticker.C {
			f()
		}
	}()
	return ticker
}

func max(d1 time.Duration, d2 time.Duration) time.Duration {
	if d1 > d2 {
		return d1
	}
	return d2
}
