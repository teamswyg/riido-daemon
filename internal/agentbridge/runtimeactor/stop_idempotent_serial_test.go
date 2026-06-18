package runtimeactor

import (
	"context"
	"testing"
	"time"
)

func TestStopIsIdempotentSerially(t *testing.T) {
	actor := startStoppableActor(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := actor.Stop(ctx); err != nil {
		t.Fatalf("first stop: %v", err)
	}
	if err := actor.Stop(ctx); err != nil {
		t.Fatalf("second stop: %v", err)
	}
	if err := actor.Stop(ctx); err != nil {
		t.Fatalf("third stop: %v", err)
	}
}
