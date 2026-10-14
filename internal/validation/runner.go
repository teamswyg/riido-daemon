package validation

import (
	"context"
	"time"
)

func RunCommand(ctx context.Context, req CommandRequest, now time.Time) (CommandResult, error) {
	normalized, err := normalizeCommandRequest(req)
	if err != nil {
		return CommandResult{}, err
	}
	started := validationStartTime(now)
	execution := executeValidationCommand(ctx, normalized)
	finished := time.Now().UTC()
	return buildCommandResult(req, normalized, execution, started, finished), nil
}
