package runtimeactor

import (
	"context"
	"testing"
	"time"
)

func startActor(t *testing.T, cfg Config) (*Actor, *fakeProcess) {
	t.Helper()
	if cfg.Process == nil {
		cfg.Process = newFakeProcess()
	}
	if cfg.MaxConcurrent == 0 {
		cfg.MaxConcurrent = 2
	}
	if cfg.MailboxSize == 0 {
		cfg.MailboxSize = 8
	}
	if cfg.RuntimeID == "" {
		cfg.RuntimeID = "rt-test"
	}
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := a.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = a.Stop(ctx)
	})
	return a, cfg.Process.(*fakeProcess)
}
