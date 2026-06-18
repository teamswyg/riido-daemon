package main

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
)

func waitForReviewDemoCLIAPI(t *testing.T, socketPath string) {
	t.Helper()

	client := riidoapi.NewClient(socketPath)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var status riidoapi.Status
		if err := client.Request(context.Background(), "status", nil, &status); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("riido API did not become ready at %s", socketPath)
}

func stopReviewDemoCLIAPI(
	t *testing.T,
	cancel context.CancelFunc,
	errs <-chan error,
) {
	t.Helper()

	cancel()
	select {
	case err := <-errs:
		if err != nil {
			t.Fatalf("Serve returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for API server shutdown")
	}
}
