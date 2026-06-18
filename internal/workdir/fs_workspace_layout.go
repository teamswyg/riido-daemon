package workdir

import "path/filepath"

func workspaceForRun(root string, id TaskID, runID string) Workspace {
	taskRoot := filepath.Join(root, id.Workspace, "tasks", id.Task, "runs", runID)
	return Workspace{
		Root:         taskRoot,
		Workdir:      filepath.Join(taskRoot, "workdir"),
		Output:       filepath.Join(taskRoot, "output"),
		Logs:         filepath.Join(taskRoot, "logs"),
		Artifacts:    filepath.Join(taskRoot, "artifacts"),
		NativeConfig: filepath.Join(taskRoot, "native-config"),
		IR:           filepath.Join(taskRoot, "ir"),
	}
}
