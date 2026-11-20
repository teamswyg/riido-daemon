package supervisor

import (
	"context"
	"fmt"
	"slices"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func (a *Actor) selectNativeResume(ctx context.Context, rt *runtimeactor.Actor, status *runtimeactor.Status, req *bridge.TaskRequest) (*runtimeactor.NativeResumeRuntime, error) {
	store, ok := a.cfg.Workdir.(workdir.NativeSessions)
	if !ok || req.Provider != "codex" || req.ResumeSessionID == "" {
		return nil, nil
	}
	record, found, err := store.LookupNativeSession("codex", req.ResumeSessionID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid local binding", errNativeSessionBinding)
	}
	if !found {
		return nil, nil
	}
	idx := slices.IndexFunc(status.Capabilities, func(c runtimeactor.Capability) bool { return c.Provider == "codex" })
	if idx < 0 || status.Capabilities[idx].CapabilityFingerprint == record.Fingerprint {
		return nil, nil
	}
	// The candidate search cannot grant a different task access to session history.
	resumeCapability := status.Capabilities[idx]
	resumeCapability.CapabilityFingerprint = record.Fingerprint
	want, err := a.nativeSessionRecord(req, resumeCapability, req.ResumeSessionID)
	if err != nil {
		return nil, err
	}
	hint := record.RuntimeHint
	record.RuntimeHint = workdir.NativeRuntimeHint{}
	record.Workdir = ""
	want.OriginRunID = record.OriginRunID
	if want != record {
		return nil, fmt.Errorf("%w: task, agent or repository grant changed", errNativeSessionBinding)
	}
	selected, err := rt.ResolveNativeResume(ctx, "codex", record.Fingerprint, hint.Executable, hint.Version)
	if err != nil {
		return nil, fmt.Errorf("%w: original Codex CLI unavailable: %w", errNativeSessionBinding, err)
	}
	status.Capabilities = slices.Clone(status.Capabilities)
	status.Capabilities[idx] = selected.Capability()
	return selected, nil
}
