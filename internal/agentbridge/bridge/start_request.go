package bridge

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

func newStartRequest(req TaskRequest, executable string) (agentbridge.StartRequest, map[string]string) {
	launchEnv := detectutil.EnvMapWithLaunchPATH(req.Env)
	return agentbridge.StartRequest{
		TaskID:          req.ID,
		Prompt:          req.Prompt,
		Cwd:             req.Cwd,
		Executable:      executable,
		Model:           req.Model,
		SystemPrompt:    req.SystemPrompt,
		MaxTurns:        req.MaxTurns,
		ResumeSessionID: req.ResumeSessionID,
		Env:             launchEnv,
		CustomArgs:      req.CustomArgs,
		MCPConfig:       req.MCPConfig,
		Metadata:        req.Metadata,
	}, launchEnv
}
