package riidoapi

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestClientRequestReadHonorsTimeout(t *testing.T) {
	socketPath := shortRiidoAPISocketPath(t)
	stop := serveStalledRiidoAPI(t, socketPath)
	defer stop()

	client := NewClient(socketPath)
	client.Timeout = 50 * time.Millisecond

	var status Status
	started := time.Now()
	err := client.Request(context.Background(), string(MethodStatus), nil, &status)
	if err == nil {
		t.Fatal("Request returned nil error for stalled riido API response")
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("Request ignored timeout: elapsed=%s err=%v", elapsed, err)
	}
	if !strings.Contains(err.Error(), "read riido API response") {
		t.Fatalf("error = %q, want read evidence", err.Error())
	}
}
