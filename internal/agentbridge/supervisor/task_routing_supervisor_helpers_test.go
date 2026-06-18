package supervisor

import (
	"context"
	"testing"
	"time"
)

func startRoutingSupervisor(t *testing.T, cfg Config) *Actor {
	t.Helper()
	if cfg.PollEvery == 0 {
		cfg.PollEvery = 10 * time.Millisecond
	}
	if cfg.HeartbeatEvery == 0 {
		cfg.HeartbeatEvery = time.Hour
	}
	actor, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := actor.Start(context.Background()); err != nil {
		t.Fatalf("supervisor Start: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = actor.Stop(ctx)
	})
	return actor
}
