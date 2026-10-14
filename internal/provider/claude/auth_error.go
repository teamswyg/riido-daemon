package claude

import "strings"

const claudeAuthRecoveryMessage = "Claude Code 로그인이 필요합니다. 데스크톱 터미널에서 `claude auth status`로 상태를 확인하고 `claude auth login`으로 다시 로그인한 뒤 작업을 다시 시도해 주세요."

func normalizeClaudeErrorMessage(message string) string {
	if isClaudeAuthError(message) {
		return claudeAuthRecoveryMessage
	}
	return message
}

func isClaudeAuthError(message string) bool {
	normalized := strings.ToLower(message)
	return strings.Contains(normalized, "not logged in") ||
		strings.Contains(normalized, "please run /login") ||
		strings.Contains(normalized, "failed to authenticate") ||
		strings.Contains(normalized, "invalid authentication credentials") ||
		(strings.Contains(normalized, "401") && strings.Contains(normalized, "authentication"))
}
