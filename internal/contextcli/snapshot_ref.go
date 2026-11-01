package contextcli

import "strings"

// SnapshotRef is an opaque server reference, never an execution credential.
func SnapshotRef(raw string) (string, *CLIError) {
	ref := strings.TrimSpace(raw)
	if ref == "" || len(ref) > 256 || strings.ContainsAny(raw, "\r\n\x00") {
		return "", Fail(CodeUsage, "--since 에 유효한 snapshot ID가 필요합니다", "riido context get 이 돌려준 snapshot_id를 넘기세요")
	}
	return ref, nil
}
