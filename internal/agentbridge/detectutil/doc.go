// Package detectutil owns the small C4 helper surface concrete provider
// adapters use to implement Detect: PATH lookup with env overrides and a
// short-running version probe.
//
// It does not own provider capability classification, provider-specific
// parsing, process spawning for runs, or scheduling decisions.
package detectutil
