package controlplane

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
)

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
