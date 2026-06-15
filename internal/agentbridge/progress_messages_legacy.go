package agentbridge

import "strings"

func classifyLegacyProgressMessage(message string) (int, string, map[string]string, bool) {
	switch {
	case message == "생각 중. . ." || message == "생각 중...":
		return 1001, "agent.thinking", nil, true
	case strings.Contains(message, "지금은 다른 작업을 처리 중"):
		return 1002, "assignment.queued.agent_busy", nil, true
	case strings.Contains(message, "에이전트가 삭제되어"):
		return 1003, "assignment.stopped.agent_deleted", nil, true
	}
	if label, description, ok := splitLegacyProgress(message, " 수집 중 - "); ok {
		return 1101, "tool.collecting", map[string]string{"label": label, "description": description}, true
	}
	if label, description, ok := splitLegacyProgress(message, " 실행 중 - "); ok {
		return 1103, "tool.running", map[string]string{"label": label, "description": description}, true
	}
	if label, summary, ok := splitLegacyProgress(message, " 완료 - "); ok {
		return 1104, "tool.completed", map[string]string{"label": label, "summary": summary}, true
	}
	return 0, "", nil, false
}

func splitLegacyProgress(message, marker string) (string, string, bool) {
	before, after, ok := strings.Cut(message, marker)
	if !ok {
		return "", "", false
	}
	label := strings.TrimSpace(before)
	detail := strings.TrimSpace(after)
	if label == "" || detail == "" {
		return "", "", false
	}
	return label, detail, true
}
