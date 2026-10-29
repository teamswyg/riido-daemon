package contextcli

import (
	"regexp"
	"strings"
)

// WorkRefKind tells whether a work reference is a task key or an ID.
type WorkRefKind string

const (
	WorkRefKey WorkRefKind = "key"
	WorkRefID  WorkRefKind = "id"
)

// WorkRef identifies a Riido work item by task key (RIID-5839) or ID.
type WorkRef struct {
	Value string      `json:"value"`
	Kind  WorkRefKind `json:"kind"`
}

var (
	workKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]*-[0-9]+$`)
	workIDPattern  = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)
)

// ParseWorkRef validates a work reference without contacting any server.
func ParseWorkRef(raw string) (WorkRef, *CLIError) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return WorkRef{}, Fail(CodeUsage, "작업을 지정해야 합니다", "작업 키(RIID-5839) 또는 작업 ID를 넘기세요")
	}
	if workKeyPattern.MatchString(strings.ToUpper(value)) {
		return WorkRef{Value: strings.ToUpper(value), Kind: WorkRefKey}, nil
	}
	if workIDPattern.MatchString(value) {
		return WorkRef{Value: value, Kind: WorkRefID}, nil
	}
	return WorkRef{}, Failf(CodeUsage, "작업 키나 ID 형식이 아닙니다: %s", value)
}
