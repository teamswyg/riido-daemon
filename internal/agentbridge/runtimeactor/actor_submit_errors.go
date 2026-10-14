package runtimeactor

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func submitBuildStartError(adapter agentbridge.Adapter, err error) error {
	return fmt.Errorf("runtimeactor: BuildStart %s: %w", adapter.Name(), err)
}

func submitSessionStartError(err error) error {
	return fmt.Errorf("runtimeactor: session.Start: %w", err)
}
