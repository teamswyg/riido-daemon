package controlplane

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func (s *FileQueueSource) readQueueTask(path string) (*bridge.TaskRequest, error) {
	body, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	if err != nil {
		return nil, controlPlaneWrapf(ErrControlPlaneQueue, "file-queue.claim-task", err, "read task file")
	}
	var req bridge.TaskRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, controlPlaneWrapf(ErrControlPlaneQueue, "file-queue.claim-task", err, "parse %s", path)
	}
	return &req, nil
}
