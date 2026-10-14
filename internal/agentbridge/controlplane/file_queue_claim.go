package controlplane

import (
	"context"
	"os"
	"path/filepath"
	"sort"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func (s *FileQueueSource) ClaimTask(ctx context.Context, runtimeID string) (*bridge.TaskRequest, error) {
	if err := fileQueueContextErr(ctx); err != nil {
		return nil, err
	}
	entries, err := s.sortedQueueEntries()
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		req, claimed, err := s.claimQueueEntry(e, runtimeID)
		if err != nil || claimed {
			return req, err
		}
	}
	return nil, nil
}

func (s *FileQueueSource) sortedQueueEntries() ([]os.DirEntry, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, controlPlaneWrapf(ErrControlPlaneQueue, "file-queue.claim-task", err, "read queue dir")
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	return entries, nil
}

func taskQueueEntryPath(dir string, e os.DirEntry) (string, bool) {
	if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
		return "", false
	}
	return filepath.Join(dir, e.Name()), true
}
