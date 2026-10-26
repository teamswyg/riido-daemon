// Package contextbroker is the daemon side of the work-context CLI. It issues
// a per-run capability when a provider runtime starts, verifies it when the
// CLI calls `context_read` over local IPC, and answers within the run scope.
//
// The capability only proves "this caller belongs to an active run on this
// daemon". It never carries user authority; Riido data reads still need the
// control-plane run grant (plan v4 5-2, 5-3).
package contextbroker

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

const (
	tokenPrefix = "rrc_"
	tokenBytes  = 32
	// maxBindings bounds memory if runs end without lookups pruning them.
	maxBindings = 4096
)

// Binding is the run scope a capability stands for.
type Binding struct {
	RunKey       string    `json:"-"`
	TaskID       string    `json:"task_id,omitempty"`
	WorkspaceID  string    `json:"workspace_id,omitempty"`
	RuntimeID    string    `json:"runtime_id,omitempty"`
	AgentID      string    `json:"agent_id,omitempty"`
	AssignmentID string    `json:"assignment_id,omitempty"`
	IssuedAt     time.Time `json:"issued_at"`
}

type tokenHash [sha256.Size]byte

// Registry keeps only token hashes, so a status dump or memory read of the
// map does not reveal live capabilities.
type Registry struct {
	mu        sync.Mutex
	byHash    map[tokenHash]Binding
	hashByRun map[string]tokenHash
	now       func() time.Time
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{byHash: map[tokenHash]Binding{}, hashByRun: map[string]tokenHash{}, now: time.Now}
}

// Issue creates a capability for binding.RunKey, replacing any earlier one
// for the same run (retries and provider recovery reuse the run key).
func (r *Registry) Issue(binding Binding) (string, error) {
	if binding.RunKey == "" {
		return "", errors.New("contextbroker: run key is required")
	}
	var raw [tokenBytes]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	token := tokenPrefix + base64.RawURLEncoding.EncodeToString(raw[:])
	binding.IssuedAt = r.now()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.revokeLocked(binding.RunKey)
	if len(r.byHash) >= maxBindings {
		r.evictOldestLocked()
	}
	hash := sha256.Sum256([]byte(token))
	r.byHash[hash] = binding
	r.hashByRun[binding.RunKey] = hash
	return token, nil
}

// Lookup returns the binding for token.
func (r *Registry) Lookup(token string) (Binding, bool) {
	if token == "" {
		return Binding{}, false
	}
	hash := sha256.Sum256([]byte(token))
	r.mu.Lock()
	defer r.mu.Unlock()
	binding, ok := r.byHash[hash]
	return binding, ok
}

// Revoke drops the capability for runKey, if any.
func (r *Registry) Revoke(runKey string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.revokeLocked(runKey)
}

// Len reports how many capabilities are live.
func (r *Registry) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.byHash)
}

func (r *Registry) revokeLocked(runKey string) {
	hash, ok := r.hashByRun[runKey]
	if !ok {
		return
	}
	delete(r.byHash, hash)
	delete(r.hashByRun, runKey)
}

func (r *Registry) evictOldestLocked() {
	var oldestKey string
	var oldest time.Time
	for _, binding := range r.byHash {
		if oldestKey == "" || binding.IssuedAt.Before(oldest) {
			oldestKey, oldest = binding.RunKey, binding.IssuedAt
		}
	}
	r.revokeLocked(oldestKey)
}
