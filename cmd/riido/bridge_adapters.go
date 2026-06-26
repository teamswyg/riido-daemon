package main

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

// builtinAgentAdapters returns the canonical provider adapter set used by both
// the bridge CLI and the daemon runtime. Order is deterministic.
func builtinAgentAdapters() []agentbridge.Adapter {
	return []agentbridge.Adapter{
		bridgeClaudeAdapter{},
		bridgeCodexAdapter{},
		bridgeOpenClawAdapter{},
		bridgeCursorAdapter{},
	}
}

func builtinDaemonAgentAdapters(socketPath string) []agentbridge.Adapter {
	return []agentbridge.Adapter{
		bridgeClaudeAdapter{approvalSocket: socketPath},
		bridgeCodexAdapter{},
		bridgeOpenClawAdapter{},
		bridgeCursorAdapter{},
	}
}
