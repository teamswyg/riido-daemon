package supervisor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
	"github.com/teamswyg/riido-daemon/internal/process/childenv"
	"github.com/teamswyg/riido-daemon/internal/process/execwindow"
)

func defaultRunAssignmentGitClone(ctx context.Context, git string, args []string) error {
	cmd := exec.CommandContext(ctx, git, args...)
	execwindow.Hide(cmd)
	cmd.Env = append(detectutil.EnvListWithLaunchPATH(childenv.Environ(), ""), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
