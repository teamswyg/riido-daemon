package taskdb

import "strings"

type EvidenceResult string

const (
	EvidenceResultPassed  EvidenceResult = "passed"
	EvidenceResultFailed  EvidenceResult = "failed"
	EvidenceResultUnknown EvidenceResult = "unknown"
)

func normalizeEvidenceResult(result string, exitCode int) string {
	normalized := EvidenceResult(strings.ToLower(strings.TrimSpace(result)))
	switch normalized {
	case EvidenceResultPassed, EvidenceResultFailed, EvidenceResultUnknown:
		return string(normalized)
	case "":
		return defaultEvidenceResult(exitCode)
	default:
		return string(EvidenceResultUnknown)
	}
}

func defaultEvidenceResult(exitCode int) string {
	if exitCode == 0 {
		return string(EvidenceResultPassed)
	}
	return string(EvidenceResultFailed)
}
