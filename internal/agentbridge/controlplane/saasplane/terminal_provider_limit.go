package saasplane

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func terminalFailureIsProviderLimit(status agentbridge.ResultStatus, message string) bool {
	switch status {
	case agentbridge.ResultFailed, agentbridge.ResultBlocked:
	default:
		return false
	}
	normalized := strings.ToLower(strings.TrimSpace(message))
	if normalized == "" {
		return false
	}
	for _, marker := range providerLimitFailureMarkers() {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func providerLimitFailureMarkers() []string {
	return []string{
		"token usage limit exceeded",
		"token limit exceeded",
		"token quota exceeded",
		"usage limit exceeded",
		"quota exceeded",
		"rate limit exceeded",
		"insufficient credits",
		"insufficient credit",
		"credit limit exceeded",
		"cloud ai",
		"토큰 이용 한도 초과",
		"토큰 사용 한도 초과",
		"크레딧이 부족",
	}
}
