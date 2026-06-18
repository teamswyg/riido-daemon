// Package ingest implements the daemon-side C2 EventIngestor boundary.
//
// The ingestor is the single Append API for CanonicalEvent construction:
// callers provide a draft, the ingestor assigns event identity / schema /
// actor attribution / active daemon-policy versions, validates the scope-aware
// envelope, then writes through a Sink port.
//
// CanonicalEvent schema and envelope rules are owned by riido-contracts/ir.
// This package owns the local daemon's append-time completion, validation, and
// C7 policy redaction enforcement.
//
// This package is intentionally filesystem-free. Persistence adapters live
// outside core IR packages and implement Sink.
package ingest
