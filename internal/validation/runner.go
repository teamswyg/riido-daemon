package validation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

const (
	DefaultGate    = "deterministic-command-exit-code.v1"
	DefaultTimeout = 5 * time.Minute
)

type CommandRequest struct {
	Command        string
	Workdir        string
	Timeout        time.Duration
	CommandID      string
	Provider       string
	ValidationGate string
	Summary        string
}

type CommandResult struct {
	Command           string `json:"command"`
	Workdir           string `json:"workdir"`
	ExitCode          int    `json:"exit_code"`
	Result            string `json:"result"`
	ValidationGate    string `json:"validation_gate"`
	ProviderRunID     string `json:"provider_run_id"`
	ProviderRunResult string `json:"provider_run_result"`
	Summary           string `json:"summary"`
	StartedAt         string `json:"started_at"`
	FinishedAt        string `json:"finished_at"`
}

func RunCommand(ctx context.Context, req CommandRequest, now time.Time) (CommandResult, error) {
	command := strings.TrimSpace(req.Command)
	if command == "" {
		return CommandResult{}, errors.New("validation command is empty")
	}
	commandID := strings.TrimSpace(req.CommandID)
	if commandID == "" {
		return CommandResult{}, errors.New("validation command id is empty")
	}
	workdir := strings.TrimSpace(req.Workdir)
	if workdir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return CommandResult{}, fmt.Errorf("resolve validation workdir: %w", err)
		}
		workdir = cwd
	}
	if info, err := os.Stat(workdir); err != nil {
		return CommandResult{}, fmt.Errorf("stat validation workdir: %w", err)
	} else if !info.IsDir() {
		return CommandResult{}, fmt.Errorf("validation workdir is not a directory: %s", workdir)
	}
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	if now.IsZero() {
		now = time.Now()
	}
	started := now.UTC()
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "/bin/sh", "-lc", command)
	cmd.Dir = workdir
	output, err := cmd.CombinedOutput()
	exitCode := exitCodeFor(runCtx, err)
	result := resultForExitCode(exitCode)
	finished := time.Now().UTC()
	summary := strings.TrimSpace(req.Summary)
	if summary == "" {
		summary = summarize(command, exitCode, output, runCtx.Err())
	}
	return CommandResult{
		Command:           command,
		Workdir:           workdir,
		ExitCode:          exitCode,
		Result:            result,
		ValidationGate:    textutil.Default(req.ValidationGate, DefaultGate),
		ProviderRunID:     providerRunID(textutil.Default(req.Provider, "local"), commandID),
		ProviderRunResult: result,
		Summary:           summary,
		StartedAt:         started.Format(time.RFC3339Nano),
		FinishedAt:        finished.Format(time.RFC3339Nano),
	}, nil
}

func exitCodeFor(ctx context.Context, err error) int {
	if ctx.Err() == context.DeadlineExceeded {
		return 124
	}
	if err == nil {
		return 0
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}
	return 1
}

func resultForExitCode(exitCode int) string {
	if exitCode == 0 {
		return "passed"
	}
	return "failed"
}

func providerRunID(provider string, commandID string) string {
	return "provider-run:" + sanitizeID(provider) + ":" + sanitizeID(commandID)
}

func summarize(command string, exitCode int, output []byte, runErr error) string {
	if runErr == context.DeadlineExceeded {
		return fmt.Sprintf("validation command timed out: %s", command)
	}
	trimmed := string(bytes.TrimSpace(output))
	if len(trimmed) > 400 {
		trimmed = trimmed[:400]
	}
	if trimmed == "" {
		return fmt.Sprintf("validation command exited %d: %s", exitCode, command)
	}
	return fmt.Sprintf("validation command exited %d: %s: %s", exitCode, command, trimmed)
}

func sanitizeID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-' || r == '_' || r == '.' || r == ':':
			builder.WriteRune(r)
		default:
			builder.WriteRune('-')
		}
	}
	return builder.String()
}
