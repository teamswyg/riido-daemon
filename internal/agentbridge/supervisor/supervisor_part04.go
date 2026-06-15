package supervisor

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/scheduling"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func (a *Actor) claimOne(ctx context.Context, rt *runtimeactor.Actor, status runtimeactor.Status, inFlight map[string]*runningTask) bool {
	req, err := a.cfg.Source.ClaimTask(ctx, status.RuntimeID)
	if err != nil || req == nil {
		return false
	}
	if req.ID == "" {
		return false
	}
	if _, dup := inFlight[req.ID]; dup {
		return false
	}
	report := reportContextFor(req)
	reportCtx := controlplane.ContextWithTaskReport(ctx, report)

	_ = a.cfg.Reporter.StartTask(reportCtx, req.ID)
	eligibility := taskEligibility(status, req)
	if !eligibility.Eligible {
		_ = a.cfg.Reporter.CompleteTask(reportCtx, req.ID, agentbridge.Result{
			Status: agentbridge.ResultBlocked,
			Error:  "supervisor: runtime ineligible: " + eligibility.Summary(),
		})
		return true
	}
	prepared, err := a.prepareWorkspace(ctx, status, req)
	if err != nil {
		_ = a.cfg.Reporter.CompleteTask(reportCtx, req.ID, agentbridge.Result{
			Status: agentbridge.ResultFailed,
			Error:  err.Error(),
		})
		return true
	}
	handle, err := rt.Submit(ctx, *req)
	if err != nil {
		res := agentbridge.Result{
			Status: agentbridge.ResultFailed,
			Error:  err.Error(),
		}
		if prepared != nil {
			res = a.recordTerminalResult(ctx, &runningTask{
				taskID:    req.ID,
				report:    report,
				runtime:   rt,
				workspace: prepared.workspace,
				events:    prepared.events,
			}, res)
		}
		_ = a.cfg.Reporter.CompleteTask(reportCtx, req.ID, res)
		return true
	}
	_ = a.cfg.Reporter.ReportEvent(reportCtx, req.ID, agentbridge.Event{
		Kind:  agentbridge.EventLifecycle,
		Phase: agentbridge.StateRunning,
	})
	var ws *workdir.Workspace
	var events *workspaceEventContext
	if prepared != nil {
		ws = prepared.workspace
		events = prepared.events
	}
	inFlight[req.ID] = &runningTask{taskID: req.ID, report: report, runtime: rt, handle: handle, workspace: ws, events: events}

	go a.forwardSession(req.ID, handle.Events(), handle.Result())
	go a.forwardCancellation(ctx, req.ID)
	return true
}

func taskEligibility(status runtimeactor.Status, req *bridge.TaskRequest) scheduling.Eligibility {
	capView, ok := findCapability(status.Capabilities, string(req.Provider))
	if !ok {
		return scheduling.Eligibility{
			Eligible:  false,
			RuntimeID: capability.RuntimeID(status.RuntimeID),
			Reasons: []scheduling.IneligibilityReason{{
				Code:   "PROVIDER_NOT_REGISTERED",
				Detail: fmt.Sprintf("provider %q is not registered on runtime %q", req.Provider, status.RuntimeID),
			}},
		}
	}
	return scheduling.EvaluateCapability(taskRequirements(req), scheduling.RuntimeCapability{
		RuntimeID:                 capability.RuntimeID(status.RuntimeID),
		Provider:                  capability.ProviderKind(capView.Provider),
		CapabilityFingerprint:     capability.CapabilityFingerprint(capView.CapabilityFingerprint),
		Available:                 capView.Available,
		CompatibilityStatus:       capability.CompatibilityStatus(capView.CompatibilityStatus),
		RequiresExperimentalOptIn: capView.RequiresExperimentalOptIn,
		SupportsStreaming:         capView.SupportsStreaming,
		SupportsResume:            capView.SupportsResume,
		SupportsSystem:            capView.SupportsSystem,
		SupportsMaxTurns:          capView.SupportsMaxTurns,
		SupportsMCP:               capView.SupportsMCP,
		SupportsToolHooks:         capView.SupportsToolHooks,
		SupportsUsage:             capView.SupportsUsage,
		SupportsWorktree:          capView.SupportsWorktree,
	})
}

func reportContextFor(req *bridge.TaskRequest) controlplane.TaskReportContext {
	report, _ := controlplane.TaskReportContextFromMetadata(req.Metadata)
	return report
}

func findCapability(caps []runtimeactor.Capability, provider string) (runtimeactor.Capability, bool) {
	for _, capView := range caps {
		if capView.Provider == provider {
			return capView, true
		}
	}
	return runtimeactor.Capability{}, false
}

func runtimeTaskIDs(tasks []runtimeactor.TaskStatus) []string {
	ids := make([]string, 0, len(tasks))
	for _, task := range tasks {
		if task.TaskID != "" {
			ids = append(ids, task.TaskID)
		}
	}
	sort.Strings(ids)
	return ids
}

func taskRequirements(req *bridge.TaskRequest) scheduling.TaskRequirements {
	surfaces := make([]scheduling.RequiredSurface, 0, len(req.RequiredSurfaces))
	for _, surface := range req.RequiredSurfaces {
		surfaces = append(surfaces, scheduling.RequiredSurface(surface))
	}
	if req.Metadata != nil {
		for surface := range strings.SplitSeq(req.Metadata[MetadataRequiredSurfaces], ",") {
			surfaces = append(surfaces, scheduling.RequiredSurface(surface))
		}
	}
	return scheduling.TaskRequirements{
		Provider:                 capability.ProviderKind(req.Provider),
		RequiredSurfaces:         scheduling.NormalizeRequiredSurfaces(surfaces),
		AllowExperimentalRuntime: req.AllowExperimentalRuntime || metadataBool(req.Metadata, MetadataAllowExperimentalRuntime),
	}
}

func metadataBool(meta map[string]string, key string) bool {
	if meta == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(meta[key])) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}
