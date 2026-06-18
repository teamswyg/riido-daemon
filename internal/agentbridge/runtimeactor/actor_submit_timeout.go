package runtimeactor

import "time"

func submitHardTimeout(msg *submitMsg, fallback time.Duration) time.Duration {
	if msg.req.Timeout > 0 {
		return msg.req.Timeout
	}
	return fallback
}
