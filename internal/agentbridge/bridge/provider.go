package bridge

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

// Provider is the canonical adapter identifier (e.g. "claude", "codex").
type Provider string

// RuntimeCapability pairs a provider name with its Detect snapshot.
type RuntimeCapability struct {
	Provider Provider
	Result   agentbridge.DetectResult
}
