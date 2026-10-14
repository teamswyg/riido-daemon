package runtimeactor

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestStopIsIdempotentConcurrent(t *testing.T) {
	actor := startStoppableActor(t)
	errs := stopActorConcurrently(actor, 8)
	for i, err := range errs {
		if err != nil {
			t.Fatalf("Stop goroutine %d returned error: %v", i, err)
		}
	}
}

func stopActorConcurrently(actor *Actor, count int) []error {
	var wg sync.WaitGroup
	wg.Add(count)
	errs := make([]error, count)
	start := make(chan struct{})
	for i := range count {
		go func(idx int) {
			defer wg.Done()
			<-start
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			errs[idx] = actor.Stop(ctx)
		}(i)
	}
	close(start)
	wg.Wait()
	return errs
}
