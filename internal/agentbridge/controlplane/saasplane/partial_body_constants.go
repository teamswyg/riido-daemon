package saasplane

import (
	"time"

	"github.com/teamswyg/riido-contracts/progressmessage"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Live assistant-body streaming coalesces raw text deltas into full body
// snapshots so clients render coherent evolving assistant content.
const (
	assistantPartialProgressCode agentbridge.ProgressCode = progressmessage.AssistantPartialCode
	assistantPartialProgressKey                           = progressmessage.AssistantPartialKey
	partialBodyFlushInterval                              = 350 * time.Millisecond
	partialBodyFlushChars                                 = 24
)
