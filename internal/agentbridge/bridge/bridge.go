// Package bridge owns the C4 provider-neutral high-level entry point:
// callers register one or more provider Adapters, ask for capability
// Detect, and Run a TaskRequest to receive Events + Result.
//
// It does not own concrete provider adapters, runtime scheduling, task
// persistence, EventIngestor append authority, or local API transport.
// See docs/20-domain/provider-runtime.md §7.6.
package bridge
