// Package supervisor implements the Daemon tier of the
// Daemon -> Runtime -> Agent hierarchy.
//
// The supervisor owns the control-plane loop: register runtimes,
// heartbeat, claim tasks, submit them to the selected RuntimeActor, and report
// event/result streams back through TaskReporterPort. Its mutable state
// is owned by one goroutine; helper goroutines only translate external
// channels into mailbox messages.
package supervisor
