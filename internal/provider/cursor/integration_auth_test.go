package cursor

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func cursorAuthMissing(res agentbridge.Result, events []agentbridge.Event) bool {
	var b strings.Builder
	b.WriteString(res.Error)
	b.WriteByte(' ')
	b.WriteString(res.Output)
	for _, ev := range events {
		b.WriteByte(' ')
		b.WriteString(ev.Text)
		b.WriteByte(' ')
		b.WriteString(ev.Err)
	}
	haystack := strings.ToLower(b.String())
	return strings.Contains(haystack, "authentication required") || strings.Contains(haystack, "cursor_api_key")
}

func cursorAccountAvailable(ctx context.Context) (bool, string) {
	probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(probeCtx, DefaultExecutable, "about")
	out, _ := cmd.CombinedOutput()
	if strings.Contains(strings.ToLower(string(out)), "not logged in") {
		return false, "cursor-agent account missing; run cursor-agent login or set " + APIKeyEnv
	}
	return true, ""
}
