package main

const fixtureManifestSource = `{
  "schema_version": "riido-terminal-result-mapping.v1",
  "id": "provider-runtime-terminal-result-mapping",
  "title": "Run Lifecycle",
  "generated_doc": "docs/20-domain/provider-runtime/adapter-draft-fields/run-lifecycle.md",
  "evidence_artifact": "terminal-result-mapping-evidence",
  "sources": {
    "result_status": "internal/agentbridge/result.go",
    "terminal_result": "internal/agentbridge/supervisor/terminal_result_draft.go"
  },
  "mappings": [
    {"status": "completed", "status_const": "ResultCompleted", "event_type_const": "EventRunReportedDone", "event_type": "RunReportedDone"},
    {"status": "failed", "status_const": "ResultFailed", "event_type_const": "EventTaskFailed", "event_type": "TaskFailed"},
    {"status": "blocked", "status_const": "ResultBlocked", "event_type_const": "EventTaskFailed", "event_type": "TaskFailed"},
    {"status": "aborted", "status_const": "ResultAborted", "event_type_const": "EventTaskFailed", "event_type": "TaskFailed"},
    {"status": "cancelled", "status_const": "ResultCancelled", "event_type_const": "EventTaskCancelled", "event_type": "TaskCancelled"},
    {"status": "timeout", "status_const": "ResultTimeout", "event_type_const": "EventTaskTimedOut", "event_type": "TaskTimedOut"}
  ],
  "defaults": {
    "empty_status_const": "ResultCompleted",
    "fallback_event_type_const": "EventTaskFailed",
    "fallback_event_type": "TaskFailed"
  }
}`

const fixtureResultSource = `package agentbridge

type ResultStatus string

const (
	ResultCompleted ResultStatus = "completed"
	ResultFailed    ResultStatus = "failed"
	ResultCancelled ResultStatus = "cancelled"
	ResultTimeout   ResultStatus = "timeout"
	ResultAborted   ResultStatus = "aborted"
	ResultBlocked   ResultStatus = "blocked"
)

type Result struct {
	Status ResultStatus
}
`

const fixtureTerminalSource = `package supervisor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func terminalResultDraft(res agentbridge.Result) (string, map[string]any) {
	status := res.Status
	if status == "" {
		status = agentbridge.ResultCompleted
	}
	switch status {
	case agentbridge.ResultCompleted:
		return ir.EventRunReportedDone, nil
	case agentbridge.ResultCancelled:
		return ir.EventTaskCancelled, nil
	case agentbridge.ResultTimeout:
		return ir.EventTaskTimedOut, nil
	default:
		return ir.EventTaskFailed, nil
	}
}
`
