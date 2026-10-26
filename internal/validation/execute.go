package validation

import (
	"context"
	"os/exec"

	"github.com/teamswyg/riido-daemon/internal/process/childenv"
)

type commandExecution struct {
	Output   []byte
	ExitCode int
	Result   string
	RunErr   error
}

func executeValidationCommand(ctx context.Context, req normalizedCommandRequest) commandExecution {
	runCtx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "/bin/sh", "-lc", req.Command)
	cmd.Dir = req.Workdir
	cmd.Env = childenv.Environ()
	output, err := cmd.CombinedOutput()
	exitCode := exitCodeFor(runCtx, err)
	return commandExecution{
		Output:   output,
		ExitCode: exitCode,
		Result:   resultForExitCode(exitCode),
		RunErr:   runCtx.Err(),
	}
}
