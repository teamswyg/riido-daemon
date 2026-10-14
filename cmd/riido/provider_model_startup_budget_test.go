package main

import (
	"testing"
	"time"
)

func TestProviderRuntimeModelsStayOffSlowStartupPath(t *testing.T) {
	home := t.TempDir()
	const lookups = 200
	const budget = 250 * time.Millisecond
	start := time.Now()
	for range 100 {
		_ = cursorRuntimeModels(func() (string, error) { return home, nil })
		_ = claudeRuntimeModels()
	}
	elapsed := time.Since(start)
	if elapsed > budget {
		t.Fatalf("provider model startup path too slow: elapsed=%s budget=%s lookups=%d", elapsed, budget, lookups)
	}
	t.Logf("provider model startup path budget: elapsed=%s budget=%s lookups=%d", elapsed, budget, lookups)
}
