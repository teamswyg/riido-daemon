package main

import (
	"testing"
	"time"
)

func TestProviderRuntimeModelsStayOffSlowStartupPath(t *testing.T) {
	home := t.TempDir()
	start := time.Now()
	for range 100 {
		_ = cursorRuntimeModels(func() (string, error) { return home, nil })
		_ = claudeRuntimeModels()
	}
	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Fatalf("provider model startup path too slow: %s", elapsed)
	}
	t.Logf("provider model startup path budget: %s for 200 lookups", elapsed)
}
