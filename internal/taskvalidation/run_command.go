package taskvalidation

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/validation"
)

func runValidationCommand(
	ctx context.Context,
	req Request,
	providerForRun string,
	now time.Time,
) (validation.CommandResult, error) {
	return validation.RunCommand(ctx, validation.CommandRequest{
		Command:        req.Command,
		Workdir:        req.Workdir,
		Timeout:        req.Timeout,
		CommandID:      req.CommandID,
		Provider:       providerForRun,
		ValidationGate: req.ValidationGate,
		Summary:        req.Summary,
	}, now)
}
