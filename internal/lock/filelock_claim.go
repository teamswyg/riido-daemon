package lock

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

const (
	fileLockClaimSchemaVersion = "riido-file-lock-claim.v1"
	fileLockClaimStaleAfter    = 30 * time.Second
)

type fileLockClaim struct {
	SchemaVersion string    `json:"schema_version"`
	PID           int       `json:"pid"`
	Hostname      string    `json:"hostname,omitempty"`
	Path          string    `json:"path"`
	CreatedAt     time.Time `json:"created_at"`
	RefreshedAt   time.Time `json:"refreshed_at"`
}

func newFileLockClaim(path string, now time.Time) fileLockClaim {
	hostname, _ := os.Hostname()
	return fileLockClaim{
		SchemaVersion: fileLockClaimSchemaVersion,
		PID:           os.Getpid(),
		Hostname:      strings.TrimSpace(hostname),
		Path:          strings.TrimSpace(path),
		CreatedAt:     now.UTC(),
		RefreshedAt:   now.UTC(),
	}
}

func encodeFileLockClaim(claim fileLockClaim) ([]byte, error) {
	body, err := json.MarshalIndent(claim, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(body, '\n'), nil
}

func fileLockClaimStale(raw []byte, modTime, now time.Time) bool {
	if now.IsZero() {
		now = time.Now()
	}
	seenAt := modTime
	var claim fileLockClaim
	if err := json.Unmarshal(raw, &claim); err == nil &&
		claim.SchemaVersion == fileLockClaimSchemaVersion &&
		!claim.RefreshedAt.IsZero() {
		seenAt = claim.RefreshedAt
	}
	if seenAt.IsZero() || seenAt.After(now) {
		return false
	}
	return now.Sub(seenAt) >= fileLockClaimStaleAfter
}
