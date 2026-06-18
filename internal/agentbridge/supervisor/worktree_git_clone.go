package supervisor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

func defaultRunAssignmentGitClone(ctx context.Context, git string, args []string) error {
	cmd := exec.CommandContext(ctx, git, args...)
	cmd.Env = append(detectutil.EnvListWithLaunchPATH(os.Environ(), ""), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
