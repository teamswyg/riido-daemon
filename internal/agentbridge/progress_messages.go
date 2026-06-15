package agentbridge

import (
	"regexp"
)

const (
	ProgressMessageMetadataCode      = "riido_progress_message_code"
	ProgressMessageMetadataKey       = "riido_progress_message_key"
	ProgressMessageMetadataArgPrefix = "riido_progress_message_arg."
)

type progressMessageTemplate struct {
	Code     int
	Key      string
	Template string
}

type telemetryProgressPayload struct {
	Code    int            `json:"code"`
	Key     string         `json:"key,omitempty"`
	Args    map[string]any `json:"args,omitempty"`
	Message string         `json:"message,omitempty"`
}

var progressPlaceholderPattern = regexp.MustCompile(`\{\{([a-zA-Z0-9_]+)\}\}`)

// Projected from riido-contracts/progressmessage/catalog.ir.riido.json.
var progressMessageTemplates = []progressMessageTemplate{
	{Code: 1001, Key: "agent.thinking", Template: "생각 중. . ."},
	{Code: 1002, Key: "assignment.queued.agent_busy", Template: "지금은 다른 작업을 처리 중이에요. 현재 작업이 끝나는 대로 바로 시작할게요."},
	{Code: 1003, Key: "assignment.stopped.agent_deleted", Template: "에이전트가 삭제되어 진행 중이던 작업이 중지됐어요."},
	{Code: 1004, Key: "assignment.stopped.by_user", Template: "{{actor_name}}님이 직접 종료하였습니다"},
	{Code: 1101, Key: "tool.collecting", Template: "{{label}} 수집 중 - {{description}}"},
	{Code: 1102, Key: "tool.collection_completed_count", Template: "{{label}} 조회 완료 - {{count}}건({{representative_title}} 외)의 요약을 가져왔습니다. . ."},
	{Code: 1103, Key: "tool.running", Template: "{{label}} 실행 중 - {{description}}"},
	{Code: 1104, Key: "tool.completed", Template: "{{label}} 완료 - {{summary}}"},
	{Code: 1201, Key: "assignment.started", Template: "작업을 시작했어요."},
	{Code: 1202, Key: "assignment.completed", Template: "작업을 완료했어요."},
	{Code: 1203, Key: "assignment.failed", Template: "작업을 계속 진행하지 못했어요."},
}
