package codex

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestSharedOwnerProxyHasBoundedLifetime(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 150*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSharedOwnerProxyHelper$")
	cmd.Env = append(os.Environ(), "RIIDO_SHARED_PROXY_TEST_HELPER=1")
	started := time.Now()
	_, err := openSharedOwnerCommand(ctx, cmd)
	if err == nil || time.Since(started) > 2*time.Second {
		t.Fatal("unresponsive proxy did not terminate promptly", err)
	}
	if cmd.ProcessState == nil || cmd.ProcessState.Exited() && cmd.ProcessState.Success() {
		t.Fatal("proxy process was not reaped")
	}
}

func TestSharedOwnerProxyHelper(t *testing.T) {
	if os.Getenv("RIIDO_SHARED_PROXY_TEST_HELPER") != "1" {
		return
	}
	time.Sleep(time.Minute)
	os.Exit(0)
}
