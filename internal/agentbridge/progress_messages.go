package agentbridge

import (
	"regexp"
)

const (
	ProgressMessageMetadataCode      = "riido_progress_message_code"
	ProgressMessageMetadataKey       = "riido_progress_message_key"
	ProgressMessageMetadataArgPrefix = "riido_progress_message_arg."
)

// ProgressCode is the canonical numeric progress-message code. The underlying
// integer is preserved for telemetry metadata and DynamoDB compatibility, while
// daemon code branches on the named constants below.
type ProgressCode int

const (
	ProgressCodeUnknown ProgressCode = 0

	ProgressCodeAgentThinking                 ProgressCode = 1001
	ProgressCodeAssignmentQueuedAgentBusy     ProgressCode = 1002
	ProgressCodeAssignmentStoppedAgentDeleted ProgressCode = 1003
	ProgressCodeAssignmentStoppedByUser       ProgressCode = 1004

	ProgressCodeToolCollecting               ProgressCode = 1101
	ProgressCodeToolCollectionCompletedCount ProgressCode = 1102
	ProgressCodeToolRunning                  ProgressCode = 1103
	ProgressCodeToolCompleted                ProgressCode = 1104

	ProgressCodeAssignmentStarted   ProgressCode = 1201
	ProgressCodeAssignmentCompleted ProgressCode = 1202
	ProgressCodeAssignmentFailed    ProgressCode = 1203
)

type progressMessageTemplate struct {
	Code     ProgressCode
	Key      string
	Template string
}

type telemetryProgressPayload struct {
	Code    ProgressCode   `json:"code"`
	Key     string         `json:"key,omitempty"`
	Args    map[string]any `json:"args,omitempty"`
	Message string         `json:"message,omitempty"`
}

var progressPlaceholderPattern = regexp.MustCompile(`\{\{([a-zA-Z0-9_]+)\}\}`)

// Projected from riido-contracts/progressmessage/catalog.ir.riido.json.
var progressMessageTemplates = []progressMessageTemplate{
	{Code: ProgressCodeAgentThinking, Key: "agent.thinking", Template: "생각 중. . ."},
	{Code: ProgressCodeAssignmentQueuedAgentBusy, Key: "assignment.queued.agent_busy", Template: "지금은 다른 작업을 처리 중이에요. 현재 작업이 끝나는 대로 바로 시작할게요."},
	{Code: ProgressCodeAssignmentStoppedAgentDeleted, Key: "assignment.stopped.agent_deleted", Template: "에이전트가 삭제되어 진행 중이던 작업이 중지됐어요."},
	{Code: ProgressCodeAssignmentStoppedByUser, Key: "assignment.stopped.by_user", Template: "{{actor_name}}님이 직접 종료하였습니다"},
	{Code: ProgressCodeToolCollecting, Key: "tool.collecting", Template: "{{label}} 수집 중 - {{description}}"},
	{Code: ProgressCodeToolCollectionCompletedCount, Key: "tool.collection_completed_count", Template: "{{label}} 조회 완료 - {{count}}건({{representative_title}} 외)의 요약을 가져왔습니다. . ."},
	{Code: ProgressCodeToolRunning, Key: "tool.running", Template: "{{label}} 실행 중 - {{description}}"},
	{Code: ProgressCodeToolCompleted, Key: "tool.completed", Template: "{{label}} 완료 - {{summary}}"},
	{Code: ProgressCodeAssignmentStarted, Key: "assignment.started", Template: "작업을 시작했어요."},
	{Code: ProgressCodeAssignmentCompleted, Key: "assignment.completed", Template: "작업을 완료했어요."},
	{Code: ProgressCodeAssignmentFailed, Key: "assignment.failed", Template: "작업을 계속 진행하지 못했어요."},
}
