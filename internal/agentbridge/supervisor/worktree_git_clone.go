package supervisor

import (
	"context"
	"os/exec"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
	"github.com/teamswyg/riido-daemon/internal/process/childenv"
	"github.com/teamswyg/riido-daemon/internal/process/execwindow"
)

func defaultRunAssignmentGitClone(ctx context.Context, git string, args []string) error {
	cmd := exec.CommandContext(ctx, git, args...)
	execwindow.Hide(cmd)
	cmd.Env = append(detectutil.EnvListWithLaunchPATH(childenv.Environ(), ""),
		"GIT_TERMINAL_PROMPT=0", "GCM_INTERACTIVE=Never", "GIT_ASKPASS=", "SSH_ASKPASS=",
		"LC_ALL=C", "LANG=C")
	var diagnostic assignmentGitDiagnosticBuffer
	cmd.Stderr = &diagnostic
	// Classify stderr locally; stdout is discarded. Only fixed cause/recovery
	// messages leave this function, never Git or credential-helper output.
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return classifyAssignmentGitCloneFailure(diagnostic.data)
	}
	return nil
}
