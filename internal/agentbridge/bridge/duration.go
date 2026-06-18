package bridge

import "time"

func firstNonZero(a, b time.Duration) time.Duration {
	if a > 0 {
		return a
	}
	return b
}
