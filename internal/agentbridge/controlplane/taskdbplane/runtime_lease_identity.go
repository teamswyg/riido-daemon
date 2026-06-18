package taskdbplane

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func (r RuntimeLeaseRecord) isActive(now time.Time) bool {
	return r.ReleasedAt == nil && !r.LeaseUntil.IsZero() && !now.After(r.LeaseUntil)
}

func (r RuntimeLeaseRecord) isExpired(now time.Time) bool {
	return r.ReleasedAt == nil && !r.LeaseUntil.IsZero() && now.After(r.LeaseUntil)
}

func runtimeLeaseIndex(leases []RuntimeLeaseRecord, taskID string) int {
	for i, lease := range leases {
		if lease.TaskID == taskID {
			return i
		}
	}
	return -1
}

func runtimeLeaseID(taskID string, fencingToken int64) string {
	return fmt.Sprintf("runtime-lease:%s:%d", taskID, fencingToken)
}

func runtimeHasCapabilityFingerprint(rec controlplane.RegisteredRuntime, fingerprint string) bool {
	if strings.TrimSpace(fingerprint) == "" {
		return true
	}
	for key, value := range rec.CapabilityAttributes {
		if strings.HasPrefix(key, "provider.") && strings.HasSuffix(key, ".capability_fingerprint") && value == fingerprint {
			return true
		}
	}
	return false
}

func normalizedTaskIDs(ids []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}
