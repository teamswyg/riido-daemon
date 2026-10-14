package validation

import (
	"errors"
	"strings"
	"time"
)

type normalizedCommandRequest struct {
	Command   string
	Workdir   string
	Timeout   time.Duration
	CommandID string
}

func normalizeCommandRequest(req CommandRequest) (normalizedCommandRequest, error) {
	command := strings.TrimSpace(req.Command)
	if command == "" {
		return normalizedCommandRequest{}, errors.New("validation command is empty")
	}
	commandID := strings.TrimSpace(req.CommandID)
	if commandID == "" {
		return normalizedCommandRequest{}, errors.New("validation command id is empty")
	}
	workdir, err := resolveValidationWorkdir(req.Workdir)
	if err != nil {
		return normalizedCommandRequest{}, err
	}
	return normalizedCommandRequest{
		Command:   command,
		Workdir:   workdir,
		Timeout:   validationTimeout(req.Timeout),
		CommandID: commandID,
	}, nil
}

func validationTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return DefaultTimeout
	}
	return timeout
}
