package controlplane

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
)

func (s *FileQueueSource) moveTaskToClaim(path, runtimeID string) (string, error) {
	claimsDir := filepath.Join(s.dir, "claims")
	if err := os.MkdirAll(claimsDir, 0o755); err != nil {
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "create claims dir")
	}
	runtimeHash := sha256.Sum256([]byte(runtimeID))
	tmp, err := os.CreateTemp(claimsDir, fmt.Sprintf("%020d-%x-*.json", s.now().UTC().UnixNano(), runtimeHash[:4]))
	if err != nil {
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "reserve claim path")
	}
	claimPath := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(claimPath)
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "close claim path reservation")
	}
	if err := os.Remove(claimPath); err != nil {
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "release claim path reservation")
	}
	if err := os.Rename(path, claimPath); err != nil {
		return "", controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.move-claim", err, "rename task to claim")
	}
	return claimPath, nil
}

func (s *FileQueueSource) runtimeProviderAvailable(runtimeID, provider string) (bool, bool, error) {
	provider = strings.TrimSpace(provider)
	if runtimeID == "" || provider == "" {
		return true, false, nil
	}
	body, err := os.ReadFile(s.runtimePath(runtimeID))
	if errors.Is(err, fs.ErrNotExist) {
		return true, false, nil
	}
	if err != nil {
		return false, false, controlPlaneWrapf(ErrControlPlaneRegistry, "file-queue.runtime-provider-available", err, "read runtime registry")
	}
	rec, err := parseRuntimeRecord(body)
	if err != nil {
		return false, false, err
	}
	key := "provider." + provider + ".available"
	if available, ok := rec.Capabilities[key]; ok {
		return available, true, nil
	}
	for capabilityKey := range rec.Capabilities {
		if strings.HasPrefix(capabilityKey, "provider.") && strings.HasSuffix(capabilityKey, ".available") {
			return false, true, nil
		}
	}
	return true, false, nil
}

func (s *FileQueueSource) WatchCancellation(_ context.Context, _ string) (<-chan error, error) {
	// File queue has no out-of-band cancel channel; return a closed
	// channel so the caller can range over it without blocking.
	ch := make(chan error)
	close(ch)
	return ch, nil
}

func (s *FileQueueSource) runtimePath(runtimeID string) string {
	sum := sha256.Sum256([]byte(runtimeID))
	return filepath.Join(s.dir, "runtimes", fmt.Sprintf("%x.json", sum[:]))
}

func parseRuntimeRecord(body []byte) (RegisteredRuntime, error) {
	var rec RegisteredRuntime
	if err := json.Unmarshal(body, &rec); err != nil {
		return RegisteredRuntime{}, controlPlaneWrapf(ErrControlPlaneRegistry, "runtime-registry.parse", err, "parse runtime registry")
	}
	return rec, nil
}

func applyHeartbeat(reg *RuntimeRegistration, hb RuntimeHeartbeat) {
	if hb.RuntimeID != "" {
		reg.RuntimeID = hb.RuntimeID
	}
	if hb.DeviceName != "" {
		reg.DeviceName = hb.DeviceName
	}
	reg.UptimeSeconds = hb.UptimeSeconds
	reg.SlotLimit = hb.SlotLimit
	reg.SlotsInUse = hb.SlotsInUse
	reg.RunningTaskIDs = append([]string(nil), hb.RunningTaskIDs...)
	sort.Strings(reg.RunningTaskIDs)
}

func (s *FileQueueSource) writeRuntimeRecord(rec RegisteredRuntime) error {
	if err := fileutil.WriteJSONAtomic(s.runtimePath(rec.RuntimeID), rec); err != nil {
		return controlPlaneWrapf(ErrControlPlaneRegistry, "file-queue.write-runtime", err, "write runtime registry")
	}
	return nil
}
