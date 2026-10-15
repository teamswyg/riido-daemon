package runtimeactor

import "time"

func submitSemanticIdle(msg *submitMsg, fallback time.Duration) time.Duration {
	if msg.req.SemanticIdle > 0 {
		return msg.req.SemanticIdle
	}
	return fallback
}

func submitHardTimeout(msg *submitMsg, fallback time.Duration) time.Duration {
	if msg.req.Timeout > 0 {
		return msg.req.Timeout
	}
	return fallback
}
