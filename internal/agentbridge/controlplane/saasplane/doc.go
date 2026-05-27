// Package saasplane adapts the Riido SaaS assignment polling contract into the
// daemon's provider-neutral control-plane ports.
//
// The package owns only daemon-side HTTP client behavior for poll, heartbeat,
// event sync, and cancellation delivery. Shared assignment DTOs and state/event
// constants come from github.com/teamswyg/riido-contracts/assignment.
//
// It does not own control-plane store actors, HTTP handlers, SSE fan-out,
// request authorization, metrics/health read models, Terraform, secrets, or
// provider process execution.
package saasplane
