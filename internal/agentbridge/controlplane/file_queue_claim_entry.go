package controlplane

import (
	"errors"
	"io/fs"
	"os"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
)

func (s *FileQueueSource) claimQueueEntry(e os.DirEntry, runtimeID string) (*bridge.TaskRequest, bool, error) {
	path, ok := taskQueueEntryPath(s.dir, e)
	if !ok {
		return nil, false, nil
	}
	req, err := s.readQueueTask(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	available, ok, err := s.runtimeProviderAvailable(runtimeID, string(req.Provider))
	if err != nil || ok && !available {
		return nil, false, err
	}
	claimPath, err := s.moveTaskToClaim(path, runtimeID)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if err := s.writeClaimReceipt(claimPath, runtimeID, e.Name(), req); err != nil {
		return nil, false, err
	}
	return req, true, nil
}

func (s *FileQueueSource) writeClaimReceipt(claimPath, runtimeID, sourceFile string, req *bridge.TaskRequest) error {
	rec := FileClaimRecord{
		SchemaVersion: FileClaimRecordSchemaVersion,
		TaskID:        req.ID,
		RuntimeID:     runtimeID,
		SourceFile:    sourceFile,
		ClaimedAt:     s.now().UTC(),
		Task:          *req,
	}
	if err := fileutil.WriteJSONAtomic(claimPath, rec); err != nil {
		return controlPlaneWrapf(ErrControlPlanePersistence, "file-queue.claim-task", err, "write claim receipt")
	}
	return nil
}
