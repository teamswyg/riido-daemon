package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

const commandTimeout = 10 * time.Minute

func runStep(root, id, exe string, args ...string) stepEvidence {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, args...)
	cmd.Dir = root
	cmd.Env = localQAEnv()
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	step := stepEvidence{ID: id, Command: exe + " " + strings.Join(args, " ")}
	step.OutputTail = tail(out.String(), 4000)
	step.Status = statusPassed
	if err != nil {
		step.Status = statusFailed
		step.ExitCode = exitCode(err)
	}
	if ctx.Err() != nil {
		step.Status = statusFailed
		step.ExitCode = -1
		step.OutputTail = tail(out.String()+"\n"+ctx.Err().Error(), 4000)
	}
	return step
}

func localQAEnv() []string {
	path := "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
	return append(os.Environ(), "PATH="+path+":"+os.Getenv("PATH"))
}

func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}
