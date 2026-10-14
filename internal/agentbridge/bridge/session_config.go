package bridge

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func (c *Client) newSessionConfig(
	req TaskRequest,
	adapter agentbridge.Adapter,
	startReq agentbridge.StartRequest,
	spawnProcess process.Command,
	driver agentbridge.ProtocolDriver,
	tempFiles []string,
) session.Config {
	return session.Config{
		TaskID:               req.ID,
		RuntimeID:            string(req.Provider),
		Adapter:              adapter,
		Process:              c.process,
		Spawn:                spawnProcess,
		Request:              startReq,
		HardTimeout:          firstNonZero(req.Timeout, c.defaults.timeout),
		SemanticIdle:         firstNonZero(req.SemanticIdle, c.defaults.semanticIdle),
		AutoApprove:          c.autoApprove,
		ToolStartGate:        c.toolStartGate,
		ToolApprovalGate:     c.toolApprovalGate,
		ToolApprovalResolver: c.toolApprovalResolver,
		ProtocolDriver:       driver,
		TempFiles:            tempFiles,
	}
}
