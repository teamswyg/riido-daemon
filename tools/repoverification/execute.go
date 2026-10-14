package main

import (
	"context"
	"os"
	"os/exec"
	"time"
)

const commandTimeout = 8 * time.Minute

func runCommand(root string, spec commandSpec) commandEvidence {
	started := time.Now()
	ev := commandEvidence{ID: spec.ID, Argv: shellQuote(spec.Argv)}
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, spec.Argv[0], spec.Argv[1:]...)
	cmd.Dir = root
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	ev.DurationMS = time.Since(started).Milliseconds()
	ev.OutputTail = compactOutput(string(out))
	if err != nil {
		ev.Status = "failed"
		if ctx.Err() == context.DeadlineExceeded {
			ev.OutputTail = "timed out after " + commandTimeout.String()
		}
		return ev
	}
	ev.Status = "passed"
	return ev
}

func runCommands(root string, specs []commandSpec) []commandEvidence {
	out := make([]commandEvidence, 0, len(specs))
	for _, spec := range specs {
		out = append(out, runCommand(root, spec))
	}
	return out
}
