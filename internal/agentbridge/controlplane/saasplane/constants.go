package saasplane

import "time"

const runtimeSnapshotHeartbeatMinInterval = 4 * time.Second

const agentBindingCacheTTL = 5 * time.Second

const (
	jsonRequestMaxAttempts = 3
	jsonRequestRetryBase   = 50 * time.Millisecond
)

const (
	defaultLongPollWait       = 30 * time.Second
	longPollRequestTimeoutPad = 5 * time.Second
)
