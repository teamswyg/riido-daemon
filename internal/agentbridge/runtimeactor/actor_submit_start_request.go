package runtimeactor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

func submitLaunchEnv(msg *submitMsg) map[string]string {
	return detectutil.EnvMapWithLaunchPATH(msg.req.Env)
}

func submitStartRequest(msg *submitMsg, capView Capability, launchEnv map[string]string) agentbridge.StartRequest {
	return agentbridge.StartRequest{
		TaskID:          msg.req.ID,
		Prompt:          msg.req.Prompt,
		Cwd:             msg.req.Cwd,
		Executable:      capView.Executable,
		Model:           msg.req.Model,
		SystemPrompt:    msg.req.SystemPrompt,
		MaxTurns:        msg.req.MaxTurns,
		ResumeSessionID: msg.req.ResumeSessionID,
		Env:             launchEnv,
		CustomArgs:      msg.req.CustomArgs,
		MCPConfig:       msg.req.MCPConfig,
		Metadata:        msg.req.Metadata,
	}
}
