package supervisor

import (
	"context"
	"strings"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
	"github.com/teamswyg/riido-daemon/internal/ir/ingest"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func terminalResultDraft(res agentbridge.Result) (ir.EventType, map[string]any) {
	status := res.Status
	if status == "" {
		status = agentbridge.ResultCompleted
	}
	switch status {
	case agentbridge.ResultCompleted:
		return ir.EventRunReportedDone, map[string]any{
			"summary":      res.Output,
			"resultStatus": string(status),
		}
	case agentbridge.ResultCancelled:
		return ir.EventTaskCancelled, map[string]any{
			"reason":  textutil.FirstNonEmpty(res.Error, "provider run cancelled"),
			"byActor": "daemon",
		}
	case agentbridge.ResultTimeout:
		payload := map[string]any{
			"fromState": "Running",
			"limit":     textutil.FirstNonEmpty(res.Error, "timeout"),
		}
		if !res.StartedAt.IsZero() && !res.FinishedAt.IsZero() {
			payload["elapsed"] = res.FinishedAt.Sub(res.StartedAt).String()
		}
		return ir.EventTaskTimedOut, payload
	default:
		return ir.EventTaskFailed, map[string]any{
			"category": taskFailureCategory(status),
			"reason":   textutil.FirstNonEmpty(res.Error, string(status)),
			"terminal": true,
		}
	}
}

func taskFailureCategory(status agentbridge.ResultStatus) string {
	switch status {
	case agentbridge.ResultBlocked:
		return "provider_blocked"
	case agentbridge.ResultAborted:
		return "process_aborted"
	default:
		return "provider_result_failed"
	}
}

func toolPayload(tool agentbridge.ToolRef) map[string]any {
	payload := map[string]any{
		"toolID":   tool.ID,
		"toolName": tool.Name,
		"toolKind": tool.Kind,
		"args":     map[string]string{},
	}
	if len(tool.Args) > 0 {
		payload["args"] = toolargs.Clone(tool.Args)
	}
	return payload
}

func usagePayload(usage agentbridge.Usage) map[string]any {
	return map[string]any{
		"promptTokens":     usage.PromptTokens,
		"completionTokens": usage.CompletionTokens,
		"reasoningTokens":  usage.ReasoningTokens,
		"cacheReadTokens":  usage.CacheReadTokens,
		"cacheWriteTokens": usage.CacheWriteTokens,
	}
}

func (e *workspaceEventContext) draft(eventType ir.EventType, nativeConfigVersion string, payload map[string]any) ingest.Draft {
	return ingest.Draft{
		Scope:                 ir.EventScopeRun,
		Type:                  eventType,
		Payload:               payload,
		TaskID:                e.taskID,
		RunID:                 e.runID,
		RuntimeID:             e.runtimeID,
		CapabilityFingerprint: e.capability.CapabilityFingerprint,
		ProviderKind:          e.capability.Provider,
		ProtocolKind:          e.capability.ProtocolKind,
		ProviderVersion:       e.capability.Version,
		AdapterID:             e.capability.AdapterID,
		AdapterVersion:        e.capability.AdapterVersion,
		ProtocolVersion:       e.capability.ProtocolVersion,
		NativeConfigVersion:   nativeConfigVersion,
	}
}

func (e *workspaceEventContext) transitionDraft(eventType ir.EventType, nativeConfigVersion string, payload map[string]any) ingest.Draft {
	draft := e.draft(eventType, nativeConfigVersion, payload)
	draft.FSMVersion = task.FSMSchemaVersion
	return draft
}

func eventNativeConfigVersion(events *workspaceEventContext) string {
	if events == nil {
		return ""
	}
	return events.nativeConfigVersion
}

func firstMetadata(meta map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := meta[key]; value != "" {
			return value
		}
	}
	return ""
}

func runtimeIdentity(meta map[string]string) string {
	if value := meta[MetadataAgentIdentity]; value != "" {
		return value
	}
	if name := meta[MetadataAgentName]; name != "" {
		return "You are: " + name
	}
	return ""
}

func runtimeHardRules(meta map[string]string) []string {
	if meta == nil || strings.TrimSpace(meta[agentbridge.MetadataTelemetryContract]) == "" {
		return nil
	}
	return agentbridge.TelemetryNativeConfigHardRules()
}

func (a *Actor) forwardSession(taskID string, events <-chan agentbridge.Event, results <-chan agentbridge.Result) {
	for ev := range events {
		select {
		case a.mailbox <- envelope{taskEvent: &taskEventMsg{taskID: taskID, event: ev}}:
		case <-a.stoppedCh:
			return
		}
	}
	res, ok := <-results
	if !ok {
		return
	}
	select {
	case a.mailbox <- envelope{taskResult: &taskResultMsg{taskID: taskID, result: res}}:
	case <-a.stoppedCh:
	}
}

func (a *Actor) forwardCancellation(ctx context.Context, taskID string) {
	ch, err := a.cfg.Source.WatchCancellation(ctx, taskID)
	if err != nil {
		return
	}
	select {
	case cause, ok := <-ch:
		if !ok {
			return
		}
		select {
		case a.mailbox <- envelope{cancel: &cancelMsg{taskID: taskID, cause: cause}}:
		case <-a.stoppedCh:
		}
	case <-a.stoppedCh:
	case <-ctx.Done():
	}
}
