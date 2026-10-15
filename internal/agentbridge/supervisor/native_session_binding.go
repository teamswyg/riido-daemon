package supervisor

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/execution"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

const metadataSessionOriginRun = "native_session_origin_run"

var errNativeSessionBinding = errors.New("native session continuity could not be verified")

func (a *Actor) nativeSessionRecord(req *bridge.TaskRequest, capView runtimeactor.Capability, sessionID string) (workdir.NativeSession, error) {
	ids, err := workspacePrepareIDs(req)
	if err != nil {
		return workdir.NativeSession{}, err
	}
	binding := a.nativeSessionExecutionBinding(ids)
	data, err := json.Marshal(struct {
		Worktree any
		Binding  *execution.Binding
	}{req.Worktree, binding})
	if err != nil {
		return workdir.NativeSession{}, err
	}
	hash := sha256.Sum256(data)
	origin := firstMetadata(req.Metadata, metadataSessionOriginRun)
	if origin == "" {
		origin = ids.runID
	}
	return workdir.NativeSession{
		Schema: 1, Provider: string(req.Provider), SessionID: sessionID,
		WorkspaceID: ids.workspaceID, TaskID: ids.logicalTaskID,
		AgentID: req.Metadata["agent_id"], Fingerprint: capView.CapabilityFingerprint,
		GrantHash: hex.EncodeToString(hash[:]), OriginRunID: origin,
	}, nil
}

func (a *Actor) restoreNativeSessionWorkspace(req *bridge.TaskRequest, capView runtimeactor.Capability, ws *workdir.Workspace) (bool, error) {
	delete(req.Metadata, metadataSessionOriginRun)
	if req.ResumeSessionID == "" {
		return false, nil
	}
	if !capView.SupportsResume {
		return false, fmt.Errorf("%w: provider does not support native resume", errNativeSessionBinding)
	}
	store, ok := a.cfg.Workdir.(workdir.NativeSessions)
	if !ok {
		return false, nil
	}
	record, found, err := store.LookupNativeSession(string(req.Provider), req.ResumeSessionID)
	if err != nil {
		return false, fmt.Errorf("%w: invalid local binding", errNativeSessionBinding)
	}
	if !found {
		// Compatibility for sessions made before this index existed. Never
		// claim that their local cwd/history ownership has been verified.
		ids, idErr := workspacePrepareIDs(req)
		if idErr != nil {
			return false, idErr
		}
		if req.Worktree != nil || a.nativeSessionExecutionBinding(ids) != nil {
			return false, fmt.Errorf("%w: repository resume has no local binding", errNativeSessionBinding)
		}
		a.forwardReportEvent(taskEventMsg{taskID: req.ID, event: agentbridge.Event{
			Kind: agentbridge.EventWarning,
			Text: "Native session has no local workspace binding; using legacy workspace selection",
		}})
		return false, nil
	}
	want, err := a.nativeSessionRecord(req, capView, req.ResumeSessionID)
	if err != nil {
		return false, err
	}
	want.OriginRunID = record.OriginRunID
	cwd := record.Workdir
	record.Workdir = ""
	if want != record || !capView.SupportsResume {
		return false, fmt.Errorf("%w: task, agent, runtime or repository grant changed", errNativeSessionBinding)
	}
	// Derive the original cwd from identity, not a path supplied by the
	// provider. Resolve symlinks and reject deleted/out-of-root history.
	canonical, err := filepath.EvalSymlinks(cwd)
	if err != nil {
		return false, fmt.Errorf("%w: retained working directory is missing", errNativeSessionBinding)
	}
	runs, err := filepath.EvalSymlinks(filepath.Dir(filepath.Dir(cwd)))
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(runs, canonical)
	info, statErr := os.Stat(canonical)
	if err != nil || !filepath.IsLocal(rel) || canonical != filepath.Join(runs, record.OriginRunID, "workdir") || statErr != nil || !info.IsDir() {
		return false, fmt.Errorf("%w: retained working directory is invalid", errNativeSessionBinding)
	}
	if cloneURL, cloneErr := assignmentCloneURL(req.Worktree); cloneErr != nil {
		return false, cloneErr
	} else if cloneURL != "" {
		gitInfo, gitErr := os.Stat(filepath.Join(cwd, ".git"))
		if gitErr != nil || !gitInfo.IsDir() {
			return false, fmt.Errorf("%w: retained assignment repository is missing", errNativeSessionBinding)
		}
	}
	ws.Workdir = cwd
	req.Metadata[metadataSessionOriginRun] = record.OriginRunID
	return true, nil
}

func (a *Actor) pinTaskNativeSession(task *runningTask, ev agentbridge.Event) error {
	store, ok := a.cfg.Workdir.(workdir.NativeSessions)
	if !ok || task.request == nil || task.events == nil || !task.events.capability.SupportsResume || ev.Kind != agentbridge.EventSessionIdentified || ev.SessionID == "" {
		return nil
	}
	record, err := a.nativeSessionRecord(task.request, task.events.capability, ev.SessionID)
	if err == nil {
		err = store.PinNativeSession(record)
	}
	if err != nil {
		return fmt.Errorf("%w: could not persist provider session ownership", errNativeSessionBinding)
	}
	task.nativeSessionID = ev.SessionID
	return nil
}

func (a *Actor) nativeSessionExecutionBinding(ids workspaceIDs) *execution.Binding {
	for _, candidate := range a.cfg.ExecutionBindings {
		if candidate.WorkspaceID == ids.workspaceID && candidate.TaskID == ids.logicalTaskID {
			return &candidate
		}
	}
	return nil
}
