package runtimeactor

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
)

func (a *Actor) handleSubmit(
	adapters map[string]agentbridge.Adapter,
	caps []Capability,
	detectedAt map[string]time.Time,
	inFlight map[string]*runningTask,
	completeCh chan<- string,
	msg *submitMsg,
) (*SessionHandle, error) {
	if msg.req.ID == "" {
		return nil, errors.New("runtimeactor: TaskRequest.ID is required")
	}
	if _, dup := inFlight[msg.req.ID]; dup {
		return nil, fmt.Errorf("%w: %s", ErrDuplicateTaskID, msg.req.ID)
	}
	if len(inFlight) >= a.cfg.MaxConcurrent {
		return nil, ErrSlotExhausted
	}

	adapter, ok := adapters[string(msg.req.Provider)]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnknownProvider, msg.req.Provider)
	}
	capView, ok, err := a.capabilityForSubmit(msg.ctx, adapter, caps, detectedAt, string(msg.req.Provider))
	if err != nil {
		return nil, err
	}
	if !ok || !capView.Available {
		return nil, fmt.Errorf("%w: %s", ErrProviderUnavailable, msg.req.Provider)
	}

	launchEnv := detectutil.EnvMapWithLaunchPATH(msg.req.Env)
	startReq := agentbridge.StartRequest{
		TaskID:          msg.req.ID,
		Prompt:          msg.req.Prompt,
		Cwd:             msg.req.Cwd,
		Executable:      capView.Executable,
		Model:           msg.req.Model,
		SystemPrompt:    msg.req.SystemPrompt,
		MaxTurns:        msg.req.MaxTurns,
		ResumeSessionID: msg.req.ResumeSessionID,
		Env:             launchEnv,
		CustomArgs:      msg.req.CustomArgs,
		MCPConfig:       msg.req.MCPConfig,
		Metadata:        msg.req.Metadata,
	}
	spawn, err := adapter.BuildStart(startReq)
	if err != nil {
		return nil, fmt.Errorf("runtimeactor: BuildStart %s: %w", adapter.Name(), err)
	}

	timeout := msg.req.Timeout
	if timeout <= 0 {
		timeout = a.cfg.HardTimeout
	}
	idle := msg.req.SemanticIdle

	// Optional ProtocolDriver hook: if the adapter implements the
	// provider-neutral agentbridge.ProtocolDriverProvider port, ask it
	// for a driver to install in the session. RuntimeActor itself stays
	// generic — no provider package is imported here.
	var driver agentbridge.ProtocolDriver
	if provider, ok := adapter.(agentbridge.ProtocolDriverProvider); ok {
		drv, err := provider.NewProtocolDriver(startReq)
		if err != nil {
			return nil, fmt.Errorf("runtimeactor: NewProtocolDriver %s: %w", adapter.Name(), err)
		}
		driver = drv
	}

	spawnCommand := toProcessCommand(spawn)
	if spawnCommand.Dir == "" {
		spawnCommand.Dir = startReq.Cwd
	}
	spawnCommand.Env = detectutil.EnvListWithLaunchPATHFromMap(spawnCommand.Env, launchEnv)

	sess, err := session.Start(msg.ctx, session.Config{
		TaskID:           msg.req.ID,
		RuntimeID:        a.cfg.RuntimeID,
		Adapter:          adapter,
		Process:          a.cfg.Process,
		Spawn:            spawnCommand,
		Request:          startReq,
		HardTimeout:      timeout,
		SemanticIdle:     idle,
		AutoApprove:      a.cfg.AutoApprove,
		ToolStartGate:    a.cfg.ToolStartGate,
		ToolApprovalGate: a.cfg.ToolApprovalGate,
		ProtocolDriver:   driver,
	})
	if err != nil {
		return nil, fmt.Errorf("runtimeactor: session.Start: %w", err)
	}

	handle := &SessionHandle{TaskID: msg.req.ID, session: sess}
	task := &runningTask{
		taskID:   msg.req.ID,
		provider: string(msg.req.Provider),
		handle:   handle,
	}
	inFlight[msg.req.ID] = task

	// Watcher uses Done() so we don't consume Result.
	go func(id string, doneCh, stopped <-chan struct{}) {
		select {
		case <-doneCh:
		case <-stopped:
			return
		}
		select {
		case completeCh <- id:
		case <-stopped:
		}
	}(msg.req.ID, sess.Done(), a.stoppedCh)

	return handle, nil
}

func (a *Actor) capabilityForSubmit(
	ctx context.Context,
	adapter agentbridge.Adapter,
	caps []Capability,
	detectedAt map[string]time.Time,
	provider string,
) (Capability, bool, error) {
	idx := capabilityIndexForProvider(caps, provider)
	if idx < 0 {
		return Capability{}, false, nil
	}
	capView := caps[idx]
	if !a.capabilityRefreshDue(capView, detectedAt[provider]) {
		return capView, true, nil
	}
	refreshed, err := a.detectCapability(ctx, adapter)
	if err != nil {
		if capView.Available {
			return capView, true, nil
		}
		return Capability{}, true, fmt.Errorf("runtimeactor: Detect %s: %w", adapter.Name(), err)
	}
	caps[idx] = refreshed
	detectedAt[provider] = a.cfg.Now()
	return refreshed, true, nil
}

func (a *Actor) capabilityRefreshDue(capView Capability, detectedAt time.Time) bool {
	ttl := a.cfg.CapabilityRefreshEvery
	if ttl < 0 {
		return false
	}
	if !capView.Available && detectedAt.IsZero() {
		return true
	}
	return !detectedAt.IsZero() && a.cfg.Now().Sub(detectedAt) >= ttl
}

func (a *Actor) detectCapability(ctx context.Context, adapter agentbridge.Adapter) (Capability, error) {
	now := a.cfg.Now()
	res, err := adapter.Detect(ctx, a.cfg.DetectEnv)
	if err != nil {
		return Capability{}, err
	}
	return buildRuntimeCapability(a.cfg.RuntimeID, adapter.Name(), res, a.cfg.PolicyBundleVersion, now)
}

func (a *Actor) refreshDueCapabilities(
	ctx context.Context,
	adapters map[string]agentbridge.Adapter,
	caps []Capability,
	detectedAt map[string]time.Time,
) {
	for idx := range caps {
		capView := caps[idx]
		if !a.capabilityRefreshDue(capView, detectedAt[capView.Provider]) {
			continue
		}
		adapter, ok := adapters[capView.Provider]
		if !ok {
			continue
		}
		refreshed, err := a.detectCapability(ctx, adapter)
		if err != nil {
			continue
		}
		caps[idx] = refreshed
		detectedAt[capView.Provider] = a.cfg.Now()
	}
}

func (a *Actor) handleCancel(inFlight map[string]*runningTask, msg *cancelMsg) error {
	task, ok := inFlight[msg.taskID]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownTask, msg.taskID)
	}
	cause := errors.New(msg.reason)
	if msg.reason == "" {
		cause = errors.New("cancelled")
	}
	task.handle.session.Cancel(cause)
	return nil
}

func (a *Actor) buildStatus(caps []Capability, inFlight map[string]*runningTask) Status {
	ids := make([]string, 0, len(inFlight))
	for id := range inFlight {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	tasks := make([]TaskStatus, 0, len(ids))
	for _, id := range ids {
		t := inFlight[id]
		tasks = append(tasks, TaskStatus{
			TaskID:   t.taskID,
			Provider: t.provider,
			State:    "running",
		})
	}
	return Status{
		RuntimeID:       a.cfg.RuntimeID,
		StartedAt:       a.startedAt,
		UptimeSeconds:   int64(a.cfg.Now().Sub(a.startedAt).Seconds()),
		Health:          "ok",
		Owner:           a.cfg.Owner,
		DeviceName:      a.cfg.DeviceName,
		Agents:          append([]AgentStatus(nil), a.cfg.Agents...),
		Models:          append([]RuntimeModel(nil), a.cfg.Models...),
		Capabilities:    append([]Capability(nil), caps...),
		MaxConcurrent:   a.cfg.MaxConcurrent,
		RunningSessions: len(inFlight),
		RunningTasks:    tasks,
	}
}
