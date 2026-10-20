package supervisor

import (
	"regexp"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

var timingVersionPattern = regexp.MustCompile(`^v?[0-9]{1,4}\.[0-9]{1,4}\.[0-9]{1,6}$`)

func boundedTimingVersion(version string) string {
	if timingVersionPattern.MatchString(version) {
		return version
	}
	return "unknown"
}

func boundedTimingProvider(provider string) string {
	switch provider {
	case "claude", "codex", "cursor", "openclaw":
		return provider
	default:
		return "unknown"
	}
}

func boundedTimingStatus(status agentbridge.ResultStatus) string {
	switch status {
	case agentbridge.ResultCompleted, agentbridge.ResultFailed, agentbridge.ResultCancelled,
		agentbridge.ResultTimeout, agentbridge.ResultAborted, agentbridge.ResultBlocked:
		return string(status)
	default:
		return "unknown"
	}
}
