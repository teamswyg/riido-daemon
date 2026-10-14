package saasplane

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func toolApprovalID(tool agentbridge.ToolRef) string {
	if id := strings.TrimSpace(tool.ID); id != "" {
		return id
	}
	if id := strings.TrimSpace(tool.ProviderRequestID); id != "" {
		return id
	}
	sum := sha256.Sum256([]byte(toolApprovalIdentitySeed(tool)))
	return "approval-" + hex.EncodeToString(sum[:8])
}

func toolApprovalToolID(tool agentbridge.ToolRef, fallback string) string {
	if id := strings.TrimSpace(tool.ID); id != "" {
		return id
	}
	if id := strings.TrimSpace(tool.ProviderRequestID); id != "" {
		return id
	}
	return fallback
}

func toolApprovalDecisionReason(decision *assignmentcontract.ToolApprovalDecision) string {
	if decision == nil {
		return ""
	}
	return strings.TrimSpace(decision.Reason)
}

func toolApprovalIdentitySeed(tool agentbridge.ToolRef) string {
	return strings.TrimSpace(tool.Kind) + "\x00" +
		strings.TrimSpace(tool.Name) + "\x00" +
		time.Now().UTC().Format(time.RFC3339Nano)
}
