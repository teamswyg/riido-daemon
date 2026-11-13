package codex

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSharedOwnerProxyHasBoundedLifetime(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 150*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSharedOwnerProxyHelper$")
	cmd.Env = append(os.Environ(), "RIIDO_SHARED_PROXY_TEST_HELPER=1")
	started := time.Now()
	_, err := openSharedOwnerCommand(ctx, cmd)
	if !errors.Is(err, errSharedOwnerConnect) || time.Since(started) > 2*time.Second {
		t.Fatal("unresponsive proxy did not terminate promptly", err)
	}
	if cmd.ProcessState == nil || cmd.ProcessState.Exited() && cmd.ProcessState.Success() {
		t.Fatal("proxy process was not reaped")
	}
}

func TestSharedOwnerMissingProxyIsActionableWithoutExposingPaths(t *testing.T) {
	req := agentbridge.StartRequest{Executable: filepath.Join(t.TempDir(), "private-credential-path")}
	err := startSharedOwner(t.Context(), req, "native", "mine")
	message := sharedOwnerFailureMessage(err)
	if !errors.Is(err, errSharedOwnerConnect) || !strings.Contains(message, "codex_owner_connect_failed") || !strings.Contains(message, "종료") || strings.Contains(message, "private-credential-path") {
		t.Fatalf("missing owner diagnostic was unsafe or unactionable: %v %s", err, message)
	}
}

func TestSharedOwnerProxyHelper(t *testing.T) {
	if os.Getenv("RIIDO_SHARED_PROXY_TEST_HELPER") != "1" {
		return
	}
	time.Sleep(time.Minute)
	os.Exit(0)
}
