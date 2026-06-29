package runtimeactor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
)

func (a *Actor) startSubmitSession(
	msg *submitMsg,
	adapter agentbridge.Adapter,
	startReq agentbridge.StartRequest,
	spawn agentbridge.StartCommand,
	launchEnv map[string]string,
	driver agentbridge.ProtocolDriver,
) (*session.Session, error) {
	sess, err := session.Start(msg.ctx, session.Config{
		TaskID:               msg.req.ID,
		RuntimeID:            a.cfg.RuntimeID,
		Adapter:              adapter,
		Process:              a.cfg.Process,
		Spawn:                submitSpawnCommand(spawn, startReq, launchEnv),
		Request:              startReq,
		HardTimeout:          submitHardTimeout(msg, a.cfg.HardTimeout),
		SemanticIdle:         msg.req.SemanticIdle,
		AutoApprove:          a.cfg.AutoApprove,
		ToolStartGate:        a.cfg.ToolStartGate,
		ToolApprovalGate:     a.cfg.ToolApprovalGate,
		ToolApprovalResolver: a.cfg.ToolApprovalResolver,
		ProtocolDriver:       driver,
		TempFiles:            spawn.TempFiles,
	})
	if err != nil {
		return nil, submitSessionStartError(err)
	}
	return sess, nil
}
