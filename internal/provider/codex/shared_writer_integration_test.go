package codex

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// The fixture must be an explicitly prepared disposable chat already opened in
// the desktop app, with a previous answer RIIDO_CONTEXT_5891_PRESERVED.
func TestIntegrationDesktopWriterContinuity(t *testing.T) {
	id, cwd := os.Getenv("AGENTBRIDGE_CODEX_DESKTOP_THREAD_ID"), os.Getenv("AGENTBRIDGE_CODEX_DESKTOP_CWD")
	if os.Getenv("AGENTBRIDGE_INTEGRATION") != "1" || id == "" || cwd == "" {
		t.Skip("requires disposable desktop-owned thread fixture")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()
	req := agentbridge.StartRequest{
		TaskID: fmt.Sprintf("desktop-probe-%d", time.Now().UnixNano()), Cwd: cwd, ResumeSessionID: id,
		Prompt: "Repeat your previous final answer exactly. Do not use tools.",
	}
	res := runCodexIntegrationSession(t, ctx, req)
	if res.Status != agentbridge.ResultCompleted || res.SessionID != id || strings.TrimSpace(res.Output) != "RIIDO_CONTEXT_5891_PRESERVED" {
		t.Fatalf("desktop continuity failed: %+v", res)
	}
}
