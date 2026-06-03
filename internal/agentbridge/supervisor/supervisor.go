// Package supervisor implements the Daemon tier of the
// Daemon -> Runtime -> Agent hierarchy.
//
// The supervisor owns the control-plane loop: register runtimes,
// heartbeat, claim tasks, submit them to the selected RuntimeActor, and report
// event/result streams back through TaskReporterPort. Its mutable state
// is owned by one goroutine; helper goroutines only translate external
// channels into mailbox messages.
package supervisor

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
	"github.com/teamswyg/riido-daemon/internal/ir/ingest"
	"github.com/teamswyg/riido-daemon/internal/policy"
	"github.com/teamswyg/riido-daemon/internal/scheduling"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

var ErrStopped = errors.New("supervisor: stopped")

const (
	// DefaultMailboxSize is the supervisor actor mailbox size fixed by
	// docs/20-domain/provider-runtime.md §7.5.
	DefaultMailboxSize = 64
)

const (
	MetadataWorkspaceID              = "workspace_id"
	MetadataWorkspace                = "workspace"
	MetadataRunID                    = "run_id"
	MetadataAgentName                = "agent_name"
	MetadataAgentIdentity            = "agent_identity"
	MetadataWorkflow                 = "workflow"
	MetadataWorkdirRoot              = "workdir_root"
	MetadataWorkdir                  = "workdir"
	MetadataOutputDir                = "output_dir"
	MetadataLogsDir                  = "logs_dir"
	MetadataArtifactsDir             = "artifacts_dir"
	MetadataNativeConfig             = "native_config_dir"
	MetadataNativeConfigHome         = "native_config_home"
	MetadataIRDir                    = "ir_dir"
	MetadataNativeConfigVersion      = "native_config_version"
	MetadataRequiredSurfaces         = "required_surfaces"
	MetadataAllowExperimentalRuntime = "allow_experimental_runtime"
)

type Config struct {
	DaemonID string
	// RiidoDaemonVersion is the A-axis daemon binary version stamped on
	// CanonicalEvent common envelopes.
	RiidoDaemonVersion string
	// Runtime is the legacy single-runtime path used by tests and older
	// embedders. New daemon wiring should pass Runtimes, one RuntimeActor per
	// provider capability boundary.
	Runtime *runtimeactor.Actor
	// Runtimes is the provider-runtime pool the supervisor dispatches over.
	Runtimes []*runtimeactor.Actor
	Source   controlplane.TaskSourcePort
	Reporter controlplane.TaskReporterPort
	Workdir  workdir.Adapter

	PollEvery           time.Duration
	IdlePollEvery       time.Duration
	HeartbeatEvery      time.Duration
	MailboxSize         int
	PolicyBundleVersion string
	PolicyBundle        policy.PolicyBundle
	RuntimeTrustTier    policy.TrustTier
}

type Actor struct {
	cfg Config

	mailbox   chan envelope
	stopReqCh chan struct{}
	stoppedCh chan struct{}
	stopErrCh chan error
}

type envelope struct {
	taskEvent  *taskEventMsg
	taskResult *taskResultMsg
	cancel     *cancelMsg
}

type taskEventMsg struct {
	taskID string
	event  agentbridge.Event
}

type taskResultMsg struct {
	taskID string
	result agentbridge.Result
}

type cancelMsg struct {
	taskID string
	cause  error
}

type runningTask struct {
	taskID  string
	report  controlplane.TaskReportContext
	runtime *runtimeactor.Actor
	handle  *runtimeactor.SessionHandle

	workspace *workdir.Workspace
	events    *workspaceEventContext
}

type preparedWorkspace struct {
	workspace *workdir.Workspace
	events    *workspaceEventContext
}

type workspaceEventContext struct {
	taskID              string
	runID               string
	runtimeID           string
	capability          runtimeactor.Capability
	nativeConfigVersion string
	ingestor            *ingest.Ingestor
	agentIngestor       *ingest.Ingestor
}

func New(cfg Config) (*Actor, error) {
	if cfg.DaemonID == "" {
		return nil, errors.New("supervisor: DaemonID is required")
	}
	if len(configuredRuntimes(cfg)) == 0 {
		return nil, errors.New("supervisor: at least one Runtime is required")
	}
	if cfg.Source == nil {
		return nil, errors.New("supervisor: Source is required")
	}
	if cfg.Reporter == nil {
		return nil, errors.New("supervisor: Reporter is required")
	}
	if cfg.PollEvery <= 0 {
		cfg.PollEvery = time.Second
	}
	if cfg.IdlePollEvery <= 0 {
		cfg.IdlePollEvery = cfg.PollEvery
	}
	if cfg.IdlePollEvery < cfg.PollEvery {
		cfg.IdlePollEvery = cfg.PollEvery
	}
	if cfg.HeartbeatEvery <= 0 {
		cfg.HeartbeatEvery = 10 * time.Second
	}
	if cfg.MailboxSize <= 0 {
		cfg.MailboxSize = DefaultMailboxSize
	}
	if cfg.RiidoDaemonVersion == "" {
		cfg.RiidoDaemonVersion = "riido-agentd v0.0.0"
	}
	if cfg.PolicyBundleVersion == "" {
		cfg.PolicyBundleVersion = cfg.PolicyBundle.Version
		if cfg.PolicyBundleVersion == "" {
			cfg.PolicyBundleVersion = policy.DefaultLocalPolicyBundleVersion
		}
	}
	if cfg.PolicyBundle.SchemaVersion == "" {
		cfg.PolicyBundle = policy.DefaultLocalPolicyBundle()
		cfg.PolicyBundle.Version = cfg.PolicyBundleVersion
	} else {
		if err := cfg.PolicyBundle.Validate(); err != nil {
			return nil, fmt.Errorf("supervisor: policy bundle: %w", err)
		}
		if cfg.PolicyBundleVersion != cfg.PolicyBundle.Version {
			return nil, fmt.Errorf("supervisor: PolicyBundleVersion %q does not match policy bundle version %q", cfg.PolicyBundleVersion, cfg.PolicyBundle.Version)
		}
	}
	if cfg.RuntimeTrustTier == "" {
		cfg.RuntimeTrustTier = policy.TrustTierHost
	}
	return &Actor{
		cfg:       cfg,
		mailbox:   make(chan envelope, cfg.MailboxSize),
		stopReqCh: make(chan struct{}, 1),
		stoppedCh: make(chan struct{}),
		stopErrCh: make(chan error, 1),
	}, nil
}

func (a *Actor) Start(ctx context.Context) error {
	runtimes := configuredRuntimes(a.cfg)
	for _, rt := range runtimes {
		status, err := rt.Status(ctx)
		if err != nil {
			return fmt.Errorf("supervisor: runtime status: %w", err)
		}
		if err := a.register(ctx, status); err != nil {
			return err
		}
	}
	go a.run(ctx, runtimes)
	return nil
}

func configuredRuntimes(cfg Config) []*runtimeactor.Actor {
	if len(cfg.Runtimes) > 0 {
		out := make([]*runtimeactor.Actor, 0, len(cfg.Runtimes))
		for _, rt := range cfg.Runtimes {
			if rt != nil {
				out = append(out, rt)
			}
		}
		return out
	}
	if cfg.Runtime != nil {
		return []*runtimeactor.Actor{cfg.Runtime}
	}
	return nil
}

func (a *Actor) register(ctx context.Context, status runtimeactor.Status) error {
	caps := map[string]bool{}
	attrs := map[string]string{}
	for _, c := range status.Capabilities {
		prefix := "provider." + c.Provider + "."
		caps[prefix+"available"] = c.Available
		caps[prefix+"requires_experimental_opt_in"] = c.RequiresExperimentalOptIn
		caps[prefix+"supports_streaming"] = c.SupportsStreaming
		caps[prefix+"supports_resume"] = c.SupportsResume
		caps[prefix+"supports_system"] = c.SupportsSystem
		caps[prefix+"supports_max_turns"] = c.SupportsMaxTurns
		caps[prefix+"supports_mcp"] = c.SupportsMCP
		caps[prefix+"supports_tool_hooks"] = c.SupportsToolHooks
		caps[prefix+"supports_usage"] = c.SupportsUsage
		caps[prefix+"supports_file_events"] = c.SupportsFileEvents
		caps[prefix+"supports_worktree"] = c.SupportsWorktree
		attrs[prefix+"compatibility_status"] = c.CompatibilityStatus
		attrs[prefix+"capability_fingerprint"] = c.CapabilityFingerprint
		attrs[prefix+"protocol_kind"] = c.ProtocolKind
		attrs[prefix+"protocol_version"] = c.ProtocolVersion
		attrs[prefix+"adapter_id"] = c.AdapterID
		attrs[prefix+"adapter_version"] = c.AdapterVersion
	}
	reg := controlplane.RuntimeRegistration{
		DaemonID:             a.cfg.DaemonID,
		RuntimeID:            status.RuntimeID,
		Provider:             statusProvider(status),
		Capabilities:         caps,
		CapabilityAttributes: attrs,
		DeviceName:           status.DeviceName,
		Models:               runtimeModels(status.Models),
		StartedAt:            status.StartedAt,
		UptimeSeconds:        status.UptimeSeconds,
		SlotLimit:            status.MaxConcurrent,
		SlotsInUse:           status.RunningSessions,
		RunningTaskIDs:       runtimeTaskIDs(status.RunningTasks),
	}
	return a.cfg.Source.RegisterRuntime(ctx, reg)
}

func runtimeModels(in []runtimeactor.RuntimeModel) []controlplane.RuntimeModel {
	out := make([]controlplane.RuntimeModel, 0, len(in))
	for _, model := range in {
		out = append(out, controlplane.RuntimeModel{
			ModelID:   model.ModelID,
			Label:     model.Label,
			IsDefault: model.IsDefault,
		})
	}
	return out
}

func statusProvider(status runtimeactor.Status) string {
	if len(status.Capabilities) == 1 && status.Capabilities[0].Provider != "" {
		return status.Capabilities[0].Provider
	}
	return "multi"
}

func (a *Actor) run(ctx context.Context, runtimes []*runtimeactor.Actor) {
	defer close(a.stoppedCh)
	poll := time.NewTimer(a.cfg.PollEvery)
	defer stopTimer(poll)
	heartbeat := time.NewTicker(a.cfg.HeartbeatEvery)
	defer heartbeat.Stop()

	inFlight := map[string]*runningTask{}
	var stopErr error

	for {
		select {
		case <-ctx.Done():
			stopErr = ctx.Err()
			a.shutdown(context.Background(), runtimes, inFlight)
			a.stopErrCh <- stopErr
			return

		case <-a.stopReqCh:
			a.shutdown(context.Background(), runtimes, inFlight)
			a.stopErrCh <- nil
			return

		case <-poll.C:
			claimed := a.claimAvailable(ctx, runtimes, inFlight)
			resetTimer(poll, a.nextPollInterval(claimed, len(inFlight)))

		case <-heartbeat.C:
			for _, rt := range runtimes {
				hb, err := rt.HeartbeatPayload(ctx)
				if err != nil {
					continue
				}
				_ = a.cfg.Source.Heartbeat(ctx, controlplane.RuntimeHeartbeat{
					RuntimeID:      hb.RuntimeID,
					UptimeSeconds:  hb.UptimeSeconds,
					DeviceName:     hb.DeviceName,
					SlotLimit:      hb.SlotLimit,
					SlotsInUse:     hb.SlotsInUse,
					RunningTaskIDs: hb.RunningTaskIDs,
				})
			}

		case msg := <-a.mailbox:
			switch {
			case msg.taskEvent != nil:
				reportCtx := ctx
				if task := inFlight[msg.taskEvent.taskID]; task != nil {
					reportCtx = controlplane.ContextWithTaskReport(ctx, task.report)
					a.appendProviderEvent(ctx, msg.taskEvent.taskID, task.events, msg.taskEvent.event)
				}
				_ = a.cfg.Reporter.ReportEvent(reportCtx, msg.taskEvent.taskID, msg.taskEvent.event)
			case msg.taskResult != nil:
				running := inFlight[msg.taskResult.taskID]
				reportCtx := ctx
				res := msg.taskResult.result
				if running != nil {
					reportCtx = controlplane.ContextWithTaskReport(ctx, running.report)
					res = a.recordTerminalResult(ctx, running, msg.taskResult.result)
				}
				_ = a.cfg.Reporter.CompleteTask(reportCtx, msg.taskResult.taskID, res)
				delete(inFlight, msg.taskResult.taskID)
				resetTimer(poll, a.cfg.PollEvery)
			case msg.cancel != nil:
				if inFlight[msg.cancel.taskID] != nil {
					reason := "cancelled"
					if msg.cancel.cause != nil {
						reason = msg.cancel.cause.Error()
					}
					_ = inFlight[msg.cancel.taskID].runtime.Cancel(ctx, msg.cancel.taskID, reason)
				}
			}
		}
	}
}

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

func resetTimer(t *time.Timer, d time.Duration) {
	stopTimer(t)
	t.Reset(d)
}

func (a *Actor) nextPollInterval(claimed bool, inFlightCount int) time.Duration {
	if claimed || inFlightCount > 0 {
		return a.cfg.PollEvery
	}
	return a.cfg.IdlePollEvery
}

func (a *Actor) claimAvailable(ctx context.Context, runtimes []*runtimeactor.Actor, inFlight map[string]*runningTask) bool {
	claimed := false
	for _, rt := range runtimes {
		status, err := rt.Status(ctx)
		if err != nil {
			continue
		}
		if a.claimOne(ctx, rt, status, inFlight) {
			claimed = true
		}
	}
	return claimed
}

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
		for _, surface := range strings.Split(req.Metadata[MetadataRequiredSurfaces], ",") {
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

func (a *Actor) prepareWorkspace(ctx context.Context, status runtimeactor.Status, req *bridge.TaskRequest) (*preparedWorkspace, error) {
	if a.cfg.Workdir == nil {
		return nil, nil
	}
	if req.Metadata == nil {
		req.Metadata = map[string]string{}
	}
	workspaceID := firstMetadata(req.Metadata, MetadataWorkspaceID, MetadataWorkspace)
	if workspaceID == "" {
		return nil, errors.New("supervisor: workspace_id metadata is required when Workdir is configured")
	}
	runID := firstMetadata(req.Metadata, MetadataRunID)
	if runID == "" {
		runID = req.ID
	}
	capView, ok := findCapability(status.Capabilities, string(req.Provider))
	if !ok {
		return nil, fmt.Errorf("supervisor: capability for provider %q disappeared before workspace prepare", req.Provider)
	}
	ws, err := a.cfg.Workdir.Prepare(workdir.TaskID{Workspace: workspaceID, Task: req.ID, Run: runID})
	if err != nil {
		return nil, err
	}
	events, err := a.newWorkspaceEventContext(ws, status.RuntimeID, req, runID, capView)
	if err != nil {
		return nil, err
	}
	a.appendWorkspaceEvent(ctx, req.ID, events, ir.EventWorkdirCreated, "", map[string]any{
		"workdirPath": ws.Workdir,
		"taskID":      req.ID,
	})
	nativePlan := workdir.ProviderConfigPlan(string(req.Provider))
	nativeHookMode := a.nativeHookMode(nativePlan)
	nativeConfigHomeMode := a.nativeConfigHomeMode(nativePlan)
	resolvedNativePlan, err := workdir.ResolveProviderConfigPlanWithOptions(string(req.Provider), workdir.ProviderConfigPlanOptions{
		NativeHookMode:       nativeHookMode,
		NativeConfigHomeMode: nativeConfigHomeMode,
	})
	if err != nil {
		return nil, err
	}
	if err := a.cfg.Workdir.InjectRuntimeConfig(ws, workdir.RuntimeConfig{
		Provider:                   string(req.Provider),
		ProtocolKind:               capView.ProtocolKind,
		TelemetryContractPlacement: req.Metadata[agentbridge.MetadataTelemetryContract],
		NativeHookMode:             nativeHookMode,
		NativeConfigHomeMode:       nativeConfigHomeMode,
		Identity:                   runtimeIdentity(req.Metadata),
		HardRules:                  runtimeHardRules(req.Metadata),
		Workflow:                   req.Metadata[MetadataWorkflow],
	}); err != nil {
		return nil, err
	}
	nativeConfigVersion, err := workdir.ComputeNativeConfigVersion(ws, workdir.NativeConfigVersionInput{
		PolicyBundleVersion: a.cfg.PolicyBundleVersion,
		ProviderKind:        string(req.Provider),
		ProtocolKind:        capView.ProtocolKind,
	})
	if err != nil {
		return nil, err
	}
	req.Cwd = ws.Workdir
	req.Metadata[MetadataRunID] = runID
	req.Metadata[MetadataWorkdirRoot] = ws.Root
	req.Metadata[MetadataWorkdir] = ws.Workdir
	req.Metadata[MetadataOutputDir] = ws.Output
	req.Metadata[MetadataLogsDir] = ws.Logs
	req.Metadata[MetadataArtifactsDir] = ws.Artifacts
	req.Metadata[MetadataNativeConfig] = ws.NativeConfig
	if resolvedNativePlan.ConfigHomeDir != "" {
		req.Metadata[MetadataNativeConfigHome] = filepath.Join(ws.Workdir, filepath.FromSlash(resolvedNativePlan.ConfigHomeDir))
	} else {
		delete(req.Metadata, MetadataNativeConfigHome)
	}
	req.Metadata[MetadataIRDir] = ws.IR
	req.Metadata[MetadataNativeConfigVersion] = nativeConfigVersion
	if events != nil {
		events.nativeConfigVersion = nativeConfigVersion
	}
	a.appendWorkspaceEvent(ctx, req.ID, events, ir.EventNativeConfigInjected, nativeConfigVersion, map[string]any{
		"files":               resolvedNativePlan.GeneratedFiles(),
		"nativeConfigVersion": nativeConfigVersion,
	})
	return &preparedWorkspace{workspace: &ws, events: events}, nil
}

func (a *Actor) nativeHookMode(plan workdir.ProviderNativeConfigPlan) string {
	switch plan.HookMode {
	case workdir.NativeConfigHookModeClaudeCommandHooks:
		decision := policy.EvaluateNativeConfigHookWithBundle(a.cfg.PolicyBundle, policy.NativeConfigHookInput{
			TrustTier: a.cfg.RuntimeTrustTier,
			Surface:   policy.NativeConfigHookClaudeCommandAudit,
		})
		if decision.Allowed {
			return plan.HookMode
		}
		return workdir.NativeConfigHookModeInstructionOnly
	default:
		return plan.HookMode
	}
}

func (a *Actor) nativeConfigHomeMode(plan workdir.ProviderNativeConfigPlan) string {
	if plan.ProviderKind == "codex" && plan.ConfigHomeDir == ".codex" {
		decision := policy.EvaluateNativeConfigFileWithBundle(a.cfg.PolicyBundle, policy.NativeConfigFileInput{
			TrustTier: a.cfg.RuntimeTrustTier,
			Surface:   policy.NativeConfigFileCodexTaskScopedHome,
		})
		if decision.Allowed {
			return ""
		}
		return workdir.NativeConfigHomeModeDisabled
	}
	return ""
}

func (a *Actor) recordTerminalResult(ctx context.Context, running *runningTask, res agentbridge.Result) agentbridge.Result {
	if running == nil {
		return res
	}
	if res.Workdir == "" && running.workspace != nil {
		res.Workdir = running.workspace.Workdir
	}
	a.appendTerminalResultEvent(ctx, running.taskID, running.events, res)
	a.archiveTerminalWorkspace(ctx, running.taskID, running.workspace, running.events, res)
	return res
}

func (a *Actor) archiveTerminalWorkspace(ctx context.Context, taskID string, ws *workdir.Workspace, events *workspaceEventContext, res agentbridge.Result) {
	if ws == nil || a.cfg.Workdir == nil {
		return
	}
	archiver, ok := a.cfg.Workdir.(workdir.Archiver)
	if !ok {
		return
	}
	record, err := archiver.Archive(*ws, workdir.ArchiveRequest{
		ResultStatus: string(res.Status),
		ArchivedAt:   res.FinishedAt,
	})
	if err == nil {
		a.appendWorkspaceEvent(ctx, taskID, events, ir.EventWorkdirArchived, eventNativeConfigVersion(events), map[string]any{
			"workdirPath": record.WorkdirPath,
			"archiveURI":  record.ArchiveURI,
		})
		return
	}
	_ = a.cfg.Reporter.ReportEvent(ctx, taskID, agentbridge.Event{
		Kind: agentbridge.EventWarning,
		Text: "workspace archive failed",
		Err:  err.Error(),
	})
}

func (a *Actor) newWorkspaceEventContext(ws workdir.Workspace, statusRuntimeID string, req *bridge.TaskRequest, runID string, capView runtimeactor.Capability) (*workspaceEventContext, error) {
	sink, err := workdir.NewRunEventSink(ws)
	if err != nil {
		return nil, err
	}
	ingestor, err := ingest.New(ingest.Config{
		Sink:                sink,
		RiidoDaemonVersion:  a.cfg.RiidoDaemonVersion,
		PolicyBundleVersion: a.cfg.PolicyBundleVersion,
		ActorKind:           ir.ActorDaemon,
		ActorID:             a.cfg.DaemonID,
	})
	if err != nil {
		return nil, err
	}
	agentIngestor, err := ingest.New(ingest.Config{
		Sink:                sink,
		RiidoDaemonVersion:  a.cfg.RiidoDaemonVersion,
		PolicyBundleVersion: a.cfg.PolicyBundleVersion,
		ActorKind:           ir.ActorAgent,
		ActorID:             runID,
	})
	if err != nil {
		return nil, err
	}
	return &workspaceEventContext{
		taskID:        req.ID,
		runID:         runID,
		runtimeID:     statusRuntimeID,
		capability:    capView,
		ingestor:      ingestor,
		agentIngestor: agentIngestor,
	}, nil
}

func (a *Actor) appendWorkspaceEvent(ctx context.Context, taskID string, events *workspaceEventContext, eventType ir.EventType, nativeConfigVersion string, payload map[string]any) {
	if events == nil {
		return
	}
	if _, err := events.ingestor.Append(ctx, events.draft(eventType, nativeConfigVersion, payload)); err != nil {
		_ = a.cfg.Reporter.ReportEvent(ctx, taskID, agentbridge.Event{
			Kind: agentbridge.EventWarning,
			Text: "workspace event append failed: " + string(eventType),
			Err:  err.Error(),
		})
	}
}

func (a *Actor) appendProviderEvent(ctx context.Context, taskID string, events *workspaceEventContext, ev agentbridge.Event) {
	if events == nil {
		return
	}
	eventType, payload, ok := providerEventDraft(ev)
	if !ok {
		return
	}
	if _, err := events.agentIngestor.Append(ctx, events.draft(eventType, eventNativeConfigVersion(events), payload)); err != nil {
		_ = a.cfg.Reporter.ReportEvent(ctx, taskID, agentbridge.Event{
			Kind: agentbridge.EventWarning,
			Text: "provider event append failed: " + string(eventType),
			Err:  err.Error(),
		})
	}
}

func (a *Actor) appendTerminalResultEvent(ctx context.Context, taskID string, events *workspaceEventContext, res agentbridge.Result) {
	if events == nil {
		return
	}
	eventType, payload := terminalResultDraft(res)
	if _, err := events.ingestor.Append(ctx, events.transitionDraft(eventType, eventNativeConfigVersion(events), payload)); err != nil {
		_ = a.cfg.Reporter.ReportEvent(ctx, taskID, agentbridge.Event{
			Kind: agentbridge.EventWarning,
			Text: "terminal result event append failed: " + string(eventType),
			Err:  err.Error(),
		})
	}
}

func providerEventDraft(ev agentbridge.Event) (ir.EventType, map[string]any, bool) {
	switch ev.Kind {
	case agentbridge.EventLifecycle:
		return ir.EventStatusUpdate, map[string]any{
			"text":  "provider lifecycle update",
			"phase": string(ev.Phase),
		}, true
	case agentbridge.EventSessionIdentified:
		return ir.EventSessionPinned, map[string]any{
			"providerSessionID": ev.SessionID,
		}, true
	case agentbridge.EventTextDelta:
		return ir.EventTextDelta, map[string]any{
			"text": ev.Text,
		}, true
	case agentbridge.EventThinkingDelta:
		return ir.EventReasoningDelta, map[string]any{
			"text":    ev.Text,
			"private": true,
		}, true
	case agentbridge.EventToolCallStarted:
		return ir.EventToolCallStarted, toolPayload(ev.Tool), true
	case agentbridge.EventToolCallCompleted:
		payload := toolPayload(ev.Tool)
		payload["result"] = "completed"
		return ir.EventToolCallFinished, payload, true
	case agentbridge.EventToolCallFailed:
		payload := toolPayload(ev.Tool)
		payload["error"] = ev.Err
		return ir.EventToolCallFinished, payload, true
	case agentbridge.EventToolApprovalNeeded:
		return ir.EventApprovalRequested, map[string]any{
			"approvalID": ev.Tool.ID,
			"kind":       firstNonEmptyString(ev.Tool.Kind, "tool"),
			"payload":    toolPayload(ev.Tool),
		}, true
	case agentbridge.EventUsageDelta:
		return ir.EventUsageDelta, map[string]any{
			"usage": usagePayload(ev.Usage),
		}, true
	case agentbridge.EventLog:
		return ir.EventLogLine, map[string]any{
			"level": "info",
			"text":  ev.Text,
		}, true
	case agentbridge.EventWarning:
		return ir.EventLogLine, map[string]any{
			"level": "warning",
			"text":  ev.Text,
			"error": ev.Err,
		}, true
	case agentbridge.EventError:
		return ir.EventLogLine, map[string]any{
			"level": "error",
			"text":  ev.Text,
			"error": ev.Err,
		}, true
	default:
		return "", nil, false
	}
}

func terminalResultDraft(res agentbridge.Result) (ir.EventType, map[string]any) {
	status := res.Status
	if status == "" {
		status = agentbridge.ResultCompleted
	}
	switch status {
	case agentbridge.ResultCompleted:
		return ir.EventRunReportedDone, map[string]any{
			"summary":      res.Output,
			"resultStatus": string(status),
		}
	case agentbridge.ResultCancelled:
		return ir.EventTaskCancelled, map[string]any{
			"reason":  firstNonEmptyString(res.Error, "provider run cancelled"),
			"byActor": "daemon",
		}
	case agentbridge.ResultTimeout:
		payload := map[string]any{
			"fromState": "Running",
			"limit":     firstNonEmptyString(res.Error, "timeout"),
		}
		if !res.StartedAt.IsZero() && !res.FinishedAt.IsZero() {
			payload["elapsed"] = res.FinishedAt.Sub(res.StartedAt).String()
		}
		return ir.EventTaskTimedOut, payload
	default:
		return ir.EventTaskFailed, map[string]any{
			"category": taskFailureCategory(status),
			"reason":   firstNonEmptyString(res.Error, string(status)),
			"terminal": true,
		}
	}
}

func taskFailureCategory(status agentbridge.ResultStatus) string {
	switch status {
	case agentbridge.ResultBlocked:
		return "provider_blocked"
	case agentbridge.ResultAborted:
		return "process_aborted"
	default:
		return "provider_result_failed"
	}
}

func toolPayload(tool agentbridge.ToolRef) map[string]any {
	payload := map[string]any{
		"toolID":   tool.ID,
		"toolName": tool.Name,
		"toolKind": tool.Kind,
		"args":     map[string]string{},
	}
	if len(tool.Args) > 0 {
		payload["args"] = toolargs.Clone(tool.Args)
	}
	return payload
}

func usagePayload(usage agentbridge.Usage) map[string]any {
	return map[string]any{
		"promptTokens":     usage.PromptTokens,
		"completionTokens": usage.CompletionTokens,
		"reasoningTokens":  usage.ReasoningTokens,
		"cacheReadTokens":  usage.CacheReadTokens,
		"cacheWriteTokens": usage.CacheWriteTokens,
	}
}

func firstNonEmptyString(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func (e *workspaceEventContext) draft(eventType ir.EventType, nativeConfigVersion string, payload map[string]any) ingest.Draft {
	return ingest.Draft{
		Scope:                 ir.EventScopeRun,
		Type:                  eventType,
		Payload:               payload,
		TaskID:                e.taskID,
		RunID:                 e.runID,
		RuntimeID:             e.runtimeID,
		CapabilityFingerprint: e.capability.CapabilityFingerprint,
		ProviderKind:          e.capability.Provider,
		ProtocolKind:          e.capability.ProtocolKind,
		ProviderVersion:       e.capability.Version,
		AdapterID:             e.capability.AdapterID,
		AdapterVersion:        e.capability.AdapterVersion,
		ProtocolVersion:       e.capability.ProtocolVersion,
		NativeConfigVersion:   nativeConfigVersion,
	}
}

func (e *workspaceEventContext) transitionDraft(eventType ir.EventType, nativeConfigVersion string, payload map[string]any) ingest.Draft {
	draft := e.draft(eventType, nativeConfigVersion, payload)
	draft.FSMVersion = task.FSMSchemaVersion
	return draft
}

func eventNativeConfigVersion(events *workspaceEventContext) string {
	if events == nil {
		return ""
	}
	return events.nativeConfigVersion
}

func firstMetadata(meta map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := meta[key]; value != "" {
			return value
		}
	}
	return ""
}

func runtimeIdentity(meta map[string]string) string {
	if value := meta[MetadataAgentIdentity]; value != "" {
		return value
	}
	if name := meta[MetadataAgentName]; name != "" {
		return "You are: " + name
	}
	return ""
}

func runtimeHardRules(meta map[string]string) []string {
	if meta == nil || strings.TrimSpace(meta[agentbridge.MetadataTelemetryContract]) == "" {
		return nil
	}
	return agentbridge.TelemetryNativeConfigHardRules()
}

func (a *Actor) forwardSession(taskID string, events <-chan agentbridge.Event, results <-chan agentbridge.Result) {
	for ev := range events {
		select {
		case a.mailbox <- envelope{taskEvent: &taskEventMsg{taskID: taskID, event: ev}}:
		case <-a.stoppedCh:
			return
		}
	}
	res, ok := <-results
	if !ok {
		return
	}
	select {
	case a.mailbox <- envelope{taskResult: &taskResultMsg{taskID: taskID, result: res}}:
	case <-a.stoppedCh:
	}
}

func (a *Actor) forwardCancellation(ctx context.Context, taskID string) {
	ch, err := a.cfg.Source.WatchCancellation(ctx, taskID)
	if err != nil {
		return
	}
	select {
	case cause, ok := <-ch:
		if !ok {
			return
		}
		select {
		case a.mailbox <- envelope{cancel: &cancelMsg{taskID: taskID, cause: cause}}:
		case <-a.stoppedCh:
		}
	case <-a.stoppedCh:
	case <-ctx.Done():
	}
}

func (a *Actor) shutdown(ctx context.Context, runtimes []*runtimeactor.Actor, inFlight map[string]*runningTask) {
	finishedAt := time.Now().UTC()
	for taskID, task := range inFlight {
		_ = task.runtime.Cancel(ctx, task.taskID, ErrStopped.Error())
		res := a.recordTerminalResult(ctx, task, agentbridge.Result{
			Status:     agentbridge.ResultCancelled,
			Error:      ErrStopped.Error(),
			FinishedAt: finishedAt,
		})
		_ = a.cfg.Reporter.CompleteTask(controlplane.ContextWithTaskReport(ctx, task.report), task.taskID, res)
		delete(inFlight, taskID)
	}
	for _, rt := range runtimes {
		status, err := rt.Status(ctx)
		if err != nil || status.RuntimeID == "" {
			continue
		}
		_ = a.cfg.Source.DeregisterRuntime(ctx, status.RuntimeID)
	}
}

func (a *Actor) Stop(ctx context.Context) error {
	select {
	case <-a.stoppedCh:
		return nil
	default:
	}
	select {
	case a.stopReqCh <- struct{}{}:
	default:
	}
	select {
	case <-a.stoppedCh:
		select {
		case err := <-a.stopErrCh:
			return err
		default:
			return nil
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}
