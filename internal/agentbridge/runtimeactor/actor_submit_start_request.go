package runtimeactor

import (
	"maps"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

func submitLaunchEnv(msg *submitMsg) map[string]string {
	return detectutil.EnvMapWithLaunchPATH(msg.req.Env)
}

func submitStartRequest(msg *submitMsg, capView Capability, runtimeID string, launchEnv map[string]string) agentbridge.StartRequest {
	metadata := cloneSubmitMetadata(msg.req.Metadata)
	metadata[agentbridge.MetadataRuntimeID] = runtimeID
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
		Metadata:        metadata,
	}
}

func cloneSubmitMetadata(metadata map[string]string) map[string]string {
	out := make(map[string]string, len(metadata)+1)
	maps.Copy(out, metadata)
	return out
}
