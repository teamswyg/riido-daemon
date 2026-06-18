package runtimeactor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func indexAdapters(in []agentbridge.Adapter) map[string]agentbridge.Adapter {
	out := make(map[string]agentbridge.Adapter, len(in))
	for _, a := range in {
		out[a.Name()] = a
	}
	return out
}

func capabilityIndexForProvider(caps []Capability, provider string) int {
	for i, c := range caps {
		if c.Provider == provider {
			return i
		}
	}
	return -1
}

func metaProfile(meta map[string]string) string {
	if meta == nil {
		return ""
	}
	return meta["profile"]
}

func toProcessCommand(c agentbridge.StartCommand) process.Command {
	return process.Command{
		Executable: c.Executable,
		Args:       c.Args,
		Env:        c.Env,
		Dir:        c.Dir,
	}
}
