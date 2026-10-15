// Package providerrecovery owns finite pre-dispatch recovery fencing.
package providerrecovery

import (
	"sync"
	"time"
)

type entry struct {
	failures  int
	openUntil time.Time
	owner     string
}

// Circuits are scoped to provider plus environment fingerprint. Ownership of a
// half-open verification is atomic. No provider execution result is replayed.
type Circuits struct {
	mu     sync.Mutex
	states map[string]entry
}

func NewCircuits() *Circuits { return &Circuits{states: map[string]entry{}} }

func (c *Circuits) Acquire(scope, execution string, now time.Time) (string, time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.states[scope]
	if e.owner != "" && e.owner != execution {
		return "open", time.Second
	}
	if now.Before(e.openUntil) {
		return "open", e.openUntil.Sub(now)
	}
	if !e.openUntil.IsZero() {
		e.owner = execution
		c.states[scope] = e
		return "half_open", 0
	}
	return "closed", 0
}

func (c *Circuits) Failure(scope, execution string, now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.states[scope]
	if e.owner != "" && e.owner != execution {
		return
	}
	e.failures++
	e.owner = ""
	if e.failures >= 2 {
		e.openUntil = now.Add(time.Second)
	}
	c.setBounded(scope, e, now)
}

func (c *Circuits) Complete(scope, execution string, success bool, now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.states[scope]
	if e.owner != "" && e.owner != execution {
		return
	}
	if success {
		delete(c.states, scope)
		return
	}
	e.failures = 2
	e.owner = ""
	e.openUntil = now.Add(time.Second)
	c.setBounded(scope, e, now)
}

// Release drops half-open ownership after cancellation without classifying a
// user decision as a provider failure.
func (c *Circuits) Release(scope, execution string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.states[scope]
	if !ok || e.owner != execution {
		return
	}
	e.owner = ""
	c.states[scope] = e
}

func (c *Circuits) setBounded(scope string, e entry, now time.Time) {
	if _, exists := c.states[scope]; exists {
		c.states[scope] = e
		return
	}
	if len(c.states) >= 64 {
		for key, prior := range c.states {
			if prior.owner == "" && !now.Before(prior.openUntil) {
				delete(c.states, key)
				break
			}
		}
	}
	if len(c.states) < 64 {
		c.states[scope] = e
	}
}
