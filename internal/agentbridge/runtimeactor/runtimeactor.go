// Package runtimeactor owns the C4/C5 provider-neutral Runtime tier of
// Riido's Daemon -> Runtime -> Agent hierarchy.
//
// One Actor per local runtime capability boundary. The production daemon
// creates one RuntimeActor per provider adapter and the SupervisorActor
// dispatches across that pool. It holds:
//   - A capability snapshot for the registered provider Adapter(s).
//   - A bounded slot pool for this runtime (MaxConcurrent).
//   - The set of currently in-flight SessionActors.
//
// Actor state is owned by a single goroutine. Callers interact through
// bounded mailbox channels (Submit / Cancel / Status / Stop). No mutex
// is used in domain code. This package does not own supervisor task
// claim loops, control-plane transport, task persistence, or concrete
// provider adapters. See docs/20-domain/provider-runtime.md §7.7.
//
// The package is intentionally NOT named `runtime` to avoid colliding
// with Go's stdlib `runtime` package.
package runtimeactor

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
	"github.com/teamswyg/riido-daemon/internal/process"
)

const (
	// DefaultMailboxSize is the runtime actor mailbox size fixed by
	// docs/20-domain/provider-runtime.md §7.5.
	DefaultMailboxSize = 16
)

// ----- Errors -----

var (
	// ErrUnknownProvider — Submit referenced a provider name the Actor
	// does not have an Adapter for.
	ErrUnknownProvider = errors.New("runtimeactor: unknown provider")
	// ErrSlotExhausted — MaxConcurrent reached. Policy: reject (the
	// caller may retry or queue externally). audit M-1 §HonorsMaxConcurrentSlots.
	ErrSlotExhausted = errors.New("runtimeactor: max concurrent reached")
	// ErrUnknownTask — Cancel referenced a taskID that is not in-flight.
	ErrUnknownTask = errors.New("runtimeactor: unknown task")
	// ErrActorStopped — Submit or Cancel after Stop.
	ErrActorStopped = errors.New("runtimeactor: stopped")
	// ErrProviderUnavailable — the provider's Detect reported Available=false.
	ErrProviderUnavailable = errors.New("runtimeactor: provider unavailable")
	// ErrDuplicateTaskID — Submit with a taskID that is already running.
	ErrDuplicateTaskID = errors.New("runtimeactor: duplicate task id")
)

// ----- Public surface -----

// Capability is the daemon-side view of a single provider's runtime
// readiness. Built from Adapter.Detect.
type Capability struct {
	Provider                  string `json:"provider"`
	Available                 bool   `json:"available"`
	Version                   string `json:"version,omitempty"`
	Executable                string `json:"executable,omitempty"`
	Profile                   string `json:"profile,omitempty"`
	Reason                    string `json:"reason,omitempty"`
	ProtocolKind              string `json:"protocol_kind,omitempty"`
	AdapterID                 string `json:"adapter_id,omitempty"`
	AdapterVersion            string `json:"adapter_version,omitempty"`
	ProtocolVersion           string `json:"protocol_version,omitempty"`
	CompatibilityStatus       string `json:"compatibility_status,omitempty"`
	CapabilityFingerprint     string `json:"capability_fingerprint,omitempty"`
	DetectedFingerprint       string `json:"detected_fingerprint,omitempty"`
	RequiresExperimentalOptIn bool   `json:"requires_experimental_opt_in,omitempty"`
	SupportsStreaming         bool   `json:"supports_streaming"`
	SupportsResume            bool   `json:"supports_resume"`
	SupportsSystem            bool   `json:"supports_system"`
	SupportsMaxTurns          bool   `json:"supports_max_turns"`
	SupportsMCP               bool   `json:"supports_mcp"`
	SupportsToolHooks         bool   `json:"supports_tool_hooks"`
	SupportsUsage             bool   `json:"supports_usage"`
}

// TaskStatus describes one in-flight task within the runtime.
type TaskStatus struct {
	TaskID    string `json:"task_id"`
	Provider  string `json:"provider"`
	SessionID string `json:"session_id,omitempty"`
	State     string `json:"state"`
}

// AgentStatus is the runtime -> agent association shown by the local
// settings UI. The runtime actor does not schedule per-agent work yet;
// it simply publishes the binding data supplied by the daemon layer.
type AgentStatus struct {
	AgentID string `json:"agent_id,omitempty"`
	Name    string `json:"name"`
	State   string `json:"state,omitempty"`
}

// Status is the synchronous Status(ctx) snapshot.
type Status struct {
	RuntimeID       string        `json:"runtime_id"`
	StartedAt       time.Time     `json:"started_at"`
	UptimeSeconds   int64         `json:"uptime_seconds"`
	Health          string        `json:"health"`
	Owner           string        `json:"owner,omitempty"`
	DeviceName      string        `json:"device_name,omitempty"`
	Agents          []AgentStatus `json:"agents,omitempty"`
	Capabilities    []Capability  `json:"capabilities"`
	MaxConcurrent   int           `json:"max_concurrent"`
	RunningSessions int           `json:"running_sessions"`
	RunningTasks    []TaskStatus  `json:"running_tasks"`
}

// Heartbeat is the publish-ready payload for ControlPlane.Heartbeat.
type Heartbeat struct {
	RuntimeID      string   `json:"runtime_id"`
	UptimeSeconds  int64    `json:"uptime_seconds"`
	DeviceName     string   `json:"device_name,omitempty"`
	SlotLimit      int      `json:"slot_limit"`
	SlotsInUse     int      `json:"slots_in_use"`
	RunningTaskIDs []string `json:"running_task_ids"`
}

// SessionHandle is the caller-facing per-task handle. Mirrors
// session.Session but is the Actor's surface so we don't leak the
// internal session package across the API boundary.
type SessionHandle struct {
	TaskID  string
	session *session.Session
}

// Events returns the run-scope event stream, closed when the session
// terminates.
func (h *SessionHandle) Events() <-chan agentbridge.Event { return h.session.Events() }

// Result returns the terminal result channel (single value, then closed).
func (h *SessionHandle) Result() <-chan agentbridge.Result { return h.session.Result() }

// Done signals termination without consuming Result. Used by the Actor
// itself; callers normally prefer Result().
func (h *SessionHandle) Done() <-chan struct{} { return h.session.Done() }

// Config configures a new Actor.
type Config struct {
	RuntimeID      string
	Owner          string
	DeviceName     string
	Agents         []AgentStatus
	Adapters       []agentbridge.Adapter
	Process        process.Process
	MaxConcurrent  int
	HeartbeatEvery time.Duration // reserved for the supervisor's heartbeat loop
	HardTimeout    time.Duration // forwarded to each session as its hard timeout
	MailboxSize    int
	// AutoApprove is forwarded to each session.
	AutoApprove agentbridge.AutoApprover
	// ToolStartGate is forwarded to each session.
	ToolStartGate agentbridge.ToolStartGate
	// PolicyBundleVersion is the active policy bundle version used as a
	// CapabilityFingerprint input. Until C7 grows a policy loader, the
	// daemon supplies a Factor-12 env value and this default keeps local
	// development deterministic.
	PolicyBundleVersion string
	// DetectEnv is passed to each adapter's Detect during Start.
	DetectEnv agentbridge.DetectEnv
	// Now is injected for deterministic tests; defaults to time.Now.
	Now func() time.Time
}

// Actor is the runtime tier actor.
type Actor struct {
	cfg Config

	// Single owning goroutine writes to capabilities/inFlight/...; all
	// public methods send to mailbox channels.
	mailbox  chan envelope
	statusCh chan chan statusReply
	// stopReqCh: buffered, capacity 1. Stop callers do a non-blocking
	// send; the actor goroutine receives once. Multiple Stop calls
	// fall through the `default` branch of the send and become no-ops.
	// This replaces the previous close(stopCh)+recover() pattern,
	// eliminating the double-close panic recovery (audit H-1).
	stopReqCh chan struct{}
	stoppedCh chan struct{}
	stopErrCh chan error
	startedCh chan struct{} // closed by Start once the actor loop is live
	startedAt time.Time
}

// envelope is the actor's mailbox shape: one message kind per call.
type envelope struct {
	submit *submitMsg
	cancel *cancelMsg
}

type submitMsg struct {
	ctx   context.Context
	req   bridge.TaskRequest
	reply chan submitReply
}

type submitReply struct {
	handle *SessionHandle
	err    error
}

type cancelMsg struct {
	taskID string
	reason string
	reply  chan error
}

type statusReply struct {
	status Status
	hb     Heartbeat
}

// New validates Config and returns an Actor that has not yet started.
// Call Start(ctx) to begin the actor loop and run Detect on each adapter.
func New(cfg Config) (*Actor, error) {
	if cfg.RuntimeID == "" {
		return nil, errors.New("runtimeactor: RuntimeID is required")
	}
	if len(cfg.Adapters) == 0 {
		return nil, errors.New("runtimeactor: at least one Adapter is required")
	}
	if cfg.Process == nil {
		return nil, errors.New("runtimeactor: Process port is required")
	}
	// Detect duplicate adapter names early.
	seen := map[string]bool{}
	for _, a := range cfg.Adapters {
		if a.Name() == "" {
			return nil, errors.New("runtimeactor: adapter Name() is empty")
		}
		if seen[a.Name()] {
			return nil, fmt.Errorf("runtimeactor: duplicate adapter name %q", a.Name())
		}
		seen[a.Name()] = true
	}
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 4
	}
	if cfg.MailboxSize <= 0 {
		cfg.MailboxSize = DefaultMailboxSize
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.PolicyBundleVersion == "" {
		cfg.PolicyBundleVersion = "policy-bundle.local.v0"
	}

	a := &Actor{
		cfg:       cfg,
		mailbox:   make(chan envelope, cfg.MailboxSize),
		statusCh:  make(chan chan statusReply, 4),
		stopReqCh: make(chan struct{}, 1),
		stoppedCh: make(chan struct{}),
		stopErrCh: make(chan error, 1),
		startedCh: make(chan struct{}),
	}
	return a, nil
}

// Start runs Detect on every adapter (synchronously, on the caller's
// goroutine) and then launches the actor loop. Returns the first
// Detect error; per-provider Available=false reports do NOT abort.
func (a *Actor) Start(ctx context.Context) error {
	caps := make([]Capability, 0, len(a.cfg.Adapters))
	discoveredAt := a.cfg.Now()
	for _, adapter := range a.cfg.Adapters {
		res, err := adapter.Detect(ctx, a.cfg.DetectEnv)
		if err != nil {
			return fmt.Errorf("runtimeactor: Detect %s: %w", adapter.Name(), err)
		}
		capView, err := buildRuntimeCapability(a.cfg.RuntimeID, adapter.Name(), res, a.cfg.PolicyBundleVersion, discoveredAt)
		if err != nil {
			return fmt.Errorf("runtimeactor: capability %s: %w", adapter.Name(), err)
		}
		caps = append(caps, capView)
	}
	a.startedAt = discoveredAt
	go a.run(caps)
	close(a.startedCh)
	return nil
}

// run is the actor loop. SOLE owner of in-flight map and per-task state.
func (a *Actor) run(caps []Capability) {
	adapters := indexAdapters(a.cfg.Adapters)
	inFlight := map[string]*runningTask{}

	completeCh := make(chan string, 32)

	for {
		select {
		case env := <-a.mailbox:
			switch {
			case env.submit != nil:
				h, err := a.handleSubmit(adapters, caps, inFlight, completeCh, env.submit)
				env.submit.reply <- submitReply{handle: h, err: err}
			case env.cancel != nil:
				env.cancel.reply <- a.handleCancel(inFlight, env.cancel)
			}

		case taskID := <-completeCh:
			delete(inFlight, taskID)

		case reply := <-a.statusCh:
			reply <- statusReply{
				status: a.buildStatus(caps, inFlight),
				hb:     a.buildHeartbeat(inFlight),
			}

		case <-a.stopReqCh:
			a.drainAndShutdown(inFlight, completeCh)
			close(a.stoppedCh)
			return
		}
	}
}

// runningTask is the actor's per-task bookkeeping.
type runningTask struct {
	taskID   string
	provider string
	handle   *SessionHandle
}

func (a *Actor) handleSubmit(
	adapters map[string]agentbridge.Adapter,
	caps []Capability,
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
	if !capabilityAvailable(caps, string(msg.req.Provider)) {
		return nil, fmt.Errorf("%w: %s", ErrProviderUnavailable, msg.req.Provider)
	}

	startReq := agentbridge.StartRequest{
		TaskID:          msg.req.ID,
		Prompt:          msg.req.Prompt,
		Cwd:             msg.req.Cwd,
		Model:           msg.req.Model,
		SystemPrompt:    msg.req.SystemPrompt,
		MaxTurns:        msg.req.MaxTurns,
		ResumeSessionID: msg.req.ResumeSessionID,
		Env:             msg.req.Env,
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

	sess, err := session.Start(msg.ctx, session.Config{
		TaskID:         msg.req.ID,
		RuntimeID:      a.cfg.RuntimeID,
		Adapter:        adapter,
		Process:        a.cfg.Process,
		Spawn:          spawnCommand,
		Request:        startReq,
		HardTimeout:    timeout,
		SemanticIdle:   idle,
		AutoApprove:    a.cfg.AutoApprove,
		ToolStartGate:  a.cfg.ToolStartGate,
		ProtocolDriver: driver,
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
	go func(id string, doneCh <-chan struct{}, stopped <-chan struct{}) {
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
		Capabilities:    append([]Capability(nil), caps...),
		MaxConcurrent:   a.cfg.MaxConcurrent,
		RunningSessions: len(inFlight),
		RunningTasks:    tasks,
	}
}

func (a *Actor) buildHeartbeat(inFlight map[string]*runningTask) Heartbeat {
	ids := make([]string, 0, len(inFlight))
	for id := range inFlight {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return Heartbeat{
		RuntimeID:      a.cfg.RuntimeID,
		UptimeSeconds:  int64(a.cfg.Now().Sub(a.startedAt).Seconds()),
		DeviceName:     a.cfg.DeviceName,
		SlotLimit:      a.cfg.MaxConcurrent,
		SlotsInUse:     len(inFlight),
		RunningTaskIDs: ids,
	}
}

func (a *Actor) drainAndShutdown(inFlight map[string]*runningTask, completeCh <-chan string) {
	for _, t := range inFlight {
		t.handle.session.Cancel(ErrActorStopped)
	}
	deadline := time.After(5 * time.Second)
	for len(inFlight) > 0 {
		select {
		case id := <-completeCh:
			delete(inFlight, id)
		case <-deadline:
			a.stopErrCh <- fmt.Errorf("runtimeactor: %d session(s) did not terminate", len(inFlight))
			return
		}
	}
	a.stopErrCh <- nil
}

// ----- Public methods (mailbox-only) -----

// Submit posts a TaskRequest to the actor. Returns a SessionHandle or
// a typed error.
//
// Note on the stoppedCh check inside the reply-wait select: the
// mailbox is buffered, so a send can succeed even after Stop has fully
// shut the actor down (the actor is no longer reading). Without the
// stoppedCh guard on the wait, callers would block forever waiting
// for a reply that will never be written. The same pattern applies to
// Cancel below.
func (a *Actor) Submit(ctx context.Context, req bridge.TaskRequest) (*SessionHandle, error) {
	reply := make(chan submitReply, 1)
	select {
	case a.mailbox <- envelope{submit: &submitMsg{ctx: ctx, req: req, reply: reply}}:
	case <-a.stoppedCh:
		return nil, ErrActorStopped
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	select {
	case res := <-reply:
		return res.handle, res.err
	case <-a.stoppedCh:
		return nil, ErrActorStopped
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Cancel asks the actor to cancel an in-flight task.
func (a *Actor) Cancel(ctx context.Context, taskID, reason string) error {
	reply := make(chan error, 1)
	select {
	case a.mailbox <- envelope{cancel: &cancelMsg{taskID: taskID, reason: reason, reply: reply}}:
	case <-a.stoppedCh:
		return ErrActorStopped
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-reply:
		return err
	case <-a.stoppedCh:
		return ErrActorStopped
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Status returns a synchronous status snapshot.
func (a *Actor) Status(ctx context.Context) (Status, error) {
	reply := make(chan statusReply, 1)
	select {
	case a.statusCh <- reply:
	case <-a.stoppedCh:
		return Status{RuntimeID: a.cfg.RuntimeID, Health: "stopped"}, nil
	case <-ctx.Done():
		return Status{}, ctx.Err()
	}
	select {
	case res := <-reply:
		return res.status, nil
	case <-ctx.Done():
		return Status{}, ctx.Err()
	}
}

// HeartbeatPayload returns the publish-ready heartbeat.
func (a *Actor) HeartbeatPayload(ctx context.Context) (Heartbeat, error) {
	reply := make(chan statusReply, 1)
	select {
	case a.statusCh <- reply:
	case <-a.stoppedCh:
		return Heartbeat{RuntimeID: a.cfg.RuntimeID}, nil
	case <-ctx.Done():
		return Heartbeat{}, ctx.Err()
	}
	select {
	case res := <-reply:
		return res.hb, nil
	case <-ctx.Done():
		return Heartbeat{}, ctx.Err()
	}
}

// Stop initiates graceful shutdown. Safe to call concurrently and
// repeatedly: every Stop caller does a non-blocking send on the
// buffered stopReqCh; the first send populates the channel's single
// slot, the actor receives it once, and subsequent callers fall
// through the default branch. No channel close, no recover() panic
// guard — actor model purity (audit H-1).
func (a *Actor) Stop(ctx context.Context) error {
	select {
	case <-a.stoppedCh:
		return nil
	default:
	}
	// Non-blocking request signal. The capacity-1 buffer absorbs the
	// first sender; later senders see the slot taken and skip.
	select {
	case a.stopReqCh <- struct{}{}:
	default:
	}
	select {
	case <-a.stoppedCh:
		// stopErrCh holds the actor goroutine's exit error. Only one
		// shutdown path writes it, so the first reader gets the real
		// value; subsequent Stop callers reach this branch via
		// stoppedCh and return nil (their work is already done).
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

// ----- helpers -----

func indexAdapters(in []agentbridge.Adapter) map[string]agentbridge.Adapter {
	out := make(map[string]agentbridge.Adapter, len(in))
	for _, a := range in {
		out[a.Name()] = a
	}
	return out
}

func capabilityAvailable(caps []Capability, provider string) bool {
	for _, c := range caps {
		if c.Provider == provider {
			return c.Available
		}
	}
	return false
}

func metaProfile(meta map[string]string) string {
	if meta == nil {
		return ""
	}
	return meta["profile"]
}

func toProcessCommand(c agentbridge.StartCommand) process.Command {
	return process.Command{
		Executable: c.Executable,
		Args:       c.Args,
		Env:        c.Env,
		Dir:        c.Dir,
	}
}
