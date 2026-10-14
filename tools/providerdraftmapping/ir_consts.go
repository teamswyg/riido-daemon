package main

import "github.com/teamswyg/riido-contracts/ir"

func eventTypeValues() map[string]string {
	return map[string]string{
		"EventApprovalRequested": string(ir.EventApprovalRequested),
		"EventLogLine":           string(ir.EventLogLine),
		"EventReasoningDelta":    string(ir.EventReasoningDelta),
		"EventSessionPinned":     string(ir.EventSessionPinned),
		"EventStatusUpdate":      string(ir.EventStatusUpdate),
		"EventTextDelta":         string(ir.EventTextDelta),
		"EventToolCallFinished":  string(ir.EventToolCallFinished),
		"EventToolCallStarted":   string(ir.EventToolCallStarted),
		"EventUsageDelta":        string(ir.EventUsageDelta),
	}
}
