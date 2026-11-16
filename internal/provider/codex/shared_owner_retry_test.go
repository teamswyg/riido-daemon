package codex

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSharedOwnerConnectRetriesOnlyBeforeInitialization(t *testing.T) {
	client := &sharedOwnerClient{}
	calls := 0
	got, err := connectSharedOwner(t.Context(), func(context.Context) (*sharedOwnerClient, error) {
		calls++
		if calls < 3 {
			return nil, errSharedOwnerConnect
		}
		return client, nil
	})
	if err != nil || got != client || calls != 3 {
		t.Fatalf("transient connection not recovered: %d %v", calls, err)
	}
	for _, failure := range []error{errSharedOwnerInitialize, errSharedOwnerRejected, errSharedOwnerUnconfirmed} {
		calls = 0
		_, err := connectSharedOwner(t.Context(), func(context.Context) (*sharedOwnerClient, error) {
			calls++
			return nil, failure
		})
		if !errors.Is(err, failure) || calls != 1 {
			t.Fatal("retried after initialization or uncertain start")
		}
	}
}

func TestSharedOwnerConnectRetryIsBoundedAndCancellable(t *testing.T) {
	calls := 0
	_, err := connectSharedOwner(t.Context(), func(context.Context) (*sharedOwnerClient, error) {
		calls++
		return nil, errSharedOwnerConnect
	})
	if !errors.Is(err, errSharedOwnerConnect) || calls != 3 {
		t.Fatalf("unbounded connections: %d %v", calls, err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	calls = 0
	started := time.Now()
	_, _ = connectSharedOwner(ctx, func(context.Context) (*sharedOwnerClient, error) {
		calls++
		cancel()
		return nil, errSharedOwnerConnect
	})
	if calls != 1 || time.Since(started) > 100*time.Millisecond {
		t.Fatal("cancelled backoff retried or blocked")
	}
	_, _ = connectSharedOwner(ctx, func(context.Context) (*sharedOwnerClient, error) {
		t.Fatal("connection started after cancellation")
		return nil, nil
	})
}
