// Package runtimeactor owns the C4/C5 provider-neutral Runtime tier of
// Riido's Daemon -> Runtime -> Agent hierarchy.
//
// One Actor per local runtime capability boundary. The production daemon
// creates one RuntimeActor per provider adapter and the SupervisorActor
// dispatches across that pool. It holds:
//   - A capability snapshot for the registered provider Adapter(s).
//   - A bounded slot pool for this runtime (MaxConcurrent).
//   - The set of currently in-flight SessionActors.
//
// Actor state is owned by a single goroutine. Callers interact through
// bounded mailbox channels (Submit / Cancel / Status / Stop). No mutex
// is used in domain code. This package does not own supervisor task
// claim loops, control-plane transport, task persistence, or concrete
// provider adapters. See docs/20-domain/provider-runtime.md §7.7.
//
// The package is intentionally NOT named `runtime` to avoid colliding
// with Go's stdlib `runtime` package.
package runtimeactor
