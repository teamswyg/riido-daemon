package runtimeactor

import "time"

const (
	// DefaultMailboxSize is the runtime actor mailbox size fixed by
	// docs/20-domain/provider-runtime.md §7.5.
	DefaultMailboxSize = 16
	// DefaultCapabilityRefreshEvery bounds stale provider detection.
	// A provider that was missing at daemon start can become available
	// without requiring a daemon restart.
	DefaultCapabilityRefreshEvery = 30 * time.Second
)
