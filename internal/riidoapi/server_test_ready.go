package riidoapi

import (
	"context"
	"testing"
	"time"
)

func waitForAPI(t *testing.T, socketPath string) {
	t.Helper()
	client := NewClient(socketPath)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var status Status
		if err := client.Request(context.Background(), "status", nil, &status); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("riido API did not become ready at %s", socketPath)
}
