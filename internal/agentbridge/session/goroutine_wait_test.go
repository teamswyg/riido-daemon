package session

import (
	"runtime"
	"time"
)

func waitNumGoroutine(target int, deadline time.Duration) int {
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		if runtime.NumGoroutine() <= target {
			return runtime.NumGoroutine()
		}
		time.Sleep(20 * time.Millisecond)
	}
	return runtime.NumGoroutine()
}
