package workdir

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Prepare creates the per-task workspace tree and writes the GC marker.
// Returns an error when workspace id is empty (security/isolation gate).
func (a *FSAdapter) Prepare(id TaskID) (Workspace, error) {
	runID, err := prepareRunID(id)
	if err != nil {
		return Workspace{}, err
	}
	ws := workspaceForRun(a.root, id, runID)
	if err := createWorkspaceDirs(ws); err != nil {
		return Workspace{}, err
	}
	if err := writeGCMeta(ws, id, runID); err != nil {
		return Workspace{}, err
	}
	return ws, nil
}

func prepareRunID(id TaskID) (string, error) {
	if strings.TrimSpace(id.Workspace) == "" {
		return "", errors.New("workdir: workspace id is required")
	}
	if strings.TrimSpace(id.Task) == "" {
		return "", errors.New("workdir: task id is required")
	}
	runID := strings.TrimSpace(id.Run)
	if runID == "" {
		runID = id.Task
	}
	if !safePathSegment(id.Workspace) || !safePathSegment(id.Task) || !safePathSegment(runID) {
		return "", errors.New("workdir: workspace, task, or run id contains a path separator or traversal")
	}
	return runID, nil
}

func createWorkspaceDirs(ws Workspace) error {
	for _, dir := range []string{ws.Workdir, ws.Output, ws.Logs, ws.Artifacts, ws.NativeConfig, ws.IR} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("workdir: mkdir %s: %w", dir, err)
		}
	}
	return nil
}
