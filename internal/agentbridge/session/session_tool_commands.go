package session

import (
	"slices"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func hasApproveToolCommand(cmds []agentbridge.Command) bool {
	return slices.ContainsFunc(cmds, func(cmd agentbridge.Command) bool {
		return cmd.Kind == agentbridge.CommandApproveTool
	})
}
