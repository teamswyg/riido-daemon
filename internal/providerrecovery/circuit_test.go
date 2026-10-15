package providerrecovery

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCircuitGrantsOnlyOneHalfOpenAndScopesEnvironment(t *testing.T) {
	c := NewCircuits()
	now := time.Now()
	c.Failure("codex:old", "initial", now)
	c.Failure("codex:old", "initial", now)
	if state, _ := c.Acquire("codex:old", "first", now); state != "open" {
		t.Fatal(state)
	}
	now = now.Add(2 * time.Second)
	var wg sync.WaitGroup
	var mu sync.Mutex
	grants := 0
	for i := range 20 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			state, _ := c.Acquire("codex:old", string(rune('a'+i)), now)
			if state == "half_open" {
				mu.Lock()
				grants++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if grants != 1 {
		t.Fatalf("half-open grants=%d", grants)
	}
	if state, _ := c.Acquire("codex:new", "new", now); state != "closed" {
		t.Fatal("new environment inherited old circuit")
	}
}

func TestCircuitCardinalityAndStaleCompletionCannotReleaseHalfOpen(t *testing.T) {
	c := NewCircuits()
	now := time.Now()
	for i := range 1000 {
		c.Complete(fmt.Sprintf("environment:%d", i), "old", false, now)
	}
	if len(c.states) > 64 {
		t.Fatalf("terminal circuit state unbounded: %d", len(c.states))
	}
	c = NewCircuits()
	c.Failure("scope", "old", now)
	c.Failure("scope", "old", now)
	now = now.Add(2 * time.Second)
	if state, _ := c.Acquire("scope", "new", now); state != "half_open" {
		t.Fatal(state)
	}
	var wg sync.WaitGroup
	for i := range 30 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i % 3 {
			case 0:
				c.Complete("scope", "old", true, now)
			case 1:
				c.Failure("scope", "old", now)
			case 2:
				c.Release("scope", "old")
			}
		}(i)
	}
	wg.Wait()
	if owner := c.states["scope"].owner; owner != "new" {
		t.Fatalf("old result released new owner: %q", owner)
	}
	c.Release("scope", "new")
	if owner := c.states["scope"].owner; owner != "" {
		t.Fatal("cancel did not release half-open ownership")
	}
	if state, _ := c.Acquire("scope", "next", now); state != "half_open" {
		t.Fatal("cancel claimed provider recovered")
	}
	c.Complete("scope", "next", true, now)
	if state, _ := c.Acquire("scope", "last", now); state != "closed" {
		t.Fatal("successful half-open did not close")
	}
}
