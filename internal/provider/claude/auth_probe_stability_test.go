package claude

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

func TestClaudeAuthProbeSeparatesStderrAndUnknownLoginState(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix probe fixture")
	}
	for _, tt := range []struct {
		stdout, stderr string
		exit           int
		want           detectutil.AuthProbeStatus
	}{
		{`{"loggedIn":true}`, "startup warning", 0, detectutil.AuthProbeAuthenticated},
		{`{}`, "", 0, detectutil.AuthProbeUnknown},
		{`{"loggedIn":null}`, "", 0, detectutil.AuthProbeUnknown},
		{`{"loggedIn":true}`, "temporary failure", 7, detectutil.AuthProbeUnknown},
		{`{"loggedIn":false}`, "", 1, detectutil.AuthProbeUnauthenticated},
	} {
		exe := filepath.Join(t.TempDir(), "claude")
		script := "#!/bin/sh\nprintf '%s\\n' '" + tt.stdout + "'\nprintf '%s\\n' '" + tt.stderr + "' >&2\nexit " + string(rune('0'+tt.exit)) + "\n"
		if err := os.WriteFile(exe, []byte(script), 0o700); err != nil {
			t.Fatal(err)
		}
		if got := claudeAuthProbe(context.Background(), exe); got != tt.want {
			t.Fatalf("probe %s = %s, want %s", tt.stdout, got, tt.want)
		}
	}
}
