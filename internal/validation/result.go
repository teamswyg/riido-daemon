package validation

import (
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func buildCommandResult(req CommandRequest, normalized normalizedCommandRequest, execution commandExecution, started, finished time.Time) CommandResult {
	summary := strings.TrimSpace(req.Summary)
	if summary == "" {
		summary = summarize(normalized.Command, execution.ExitCode, execution.Output, execution.RunErr)
	}
	return CommandResult{
		Command:           normalized.Command,
		Workdir:           normalized.Workdir,
		ExitCode:          execution.ExitCode,
		Result:            execution.Result,
		ValidationGate:    textutil.Default(req.ValidationGate, DefaultGate),
		ProviderRunID:     providerRunID(textutil.Default(req.Provider, "local"), normalized.CommandID),
		ProviderRunResult: execution.Result,
		Summary:           summary,
		StartedAt:         started.Format(time.RFC3339Nano),
		FinishedAt:        finished.Format(time.RFC3339Nano),
	}
}

func validationStartTime(now time.Time) time.Time {
	if now.IsZero() {
		now = time.Now()
	}
	return now.UTC()
}
