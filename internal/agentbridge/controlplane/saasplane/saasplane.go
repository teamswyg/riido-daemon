package saasplane

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

const (
	MetadataAssignmentID    = "riido_saas_assignment_id"
	MetadataAgentID         = "riido_saas_agent_id"
	MetadataComponentID     = "riido_saas_component_id"
	MetadataLeaseToken      = "riido_saas_lease_token"
	MetadataModelID         = "riido_saas_model_id"
	MetadataRuntimeProvider = "riido_saas_runtime_provider"
)

// AgentBinding maps a SaaS agent identity to one local provider runtime.
type AgentBinding struct {
	AgentID         string
	RuntimeProvider string
}

type Config struct {
	BaseURL        string
	DaemonID       string
	DeviceID       string
	DeviceSecret   string
	Agents         []AgentBinding
	BearerToken    string
	HTTPClient     *http.Client
	RequestTimeout time.Duration
}

type RuntimeModelRecord struct {
	ModelID   string `json:"model_id"`
	Label     string `json:"label"`
	IsDefault bool   `json:"is_default"`
}

type RuntimeSnapshotRecord struct {
	RuntimeID                 string               `json:"runtime_id"`
	Kind                      string               `json:"kind"`
	Availability              string               `json:"availability,omitempty"`
	DetectionState            string               `json:"detection_state,omitempty"`
	RequiresExperimentalOptIn bool                 `json:"requires_experimental_opt_in,omitempty"`
	Models                    []RuntimeModelRecord `json:"models,omitempty"`
}

type DeviceRuntimeSnapshotSyncRequest struct {
	DaemonID          string                  `json:"daemon_id"`
	DeviceID          string                  `json:"device_id,omitempty"`
	DeviceDisplayName string                  `json:"device_display_name,omitempty"`
	Profile           string                  `json:"profile,omitempty"`
	AppVersion        string                  `json:"app_version,omitempty"`
	Runtimes          []RuntimeSnapshotRecord `json:"runtimes"`
}

type AgentRuntimeBindingListResponse struct {
	SchemaVersion string                                   `json:"schema_version"`
	Bindings      []assignmentcontract.AgentRuntimeBinding `json:"bindings"`
}

// Plane implements both TaskSourcePort and TaskReporterPort against the
// control-plane assignment polling API. Internal state is owned by a mailbox
// goroutine so the supervisor can use the adapter without shared mutable maps.
type Plane struct {
	cfg    Config
	client *http.Client
	ops    chan stateOp
	done   chan struct{}
}

func New(cfg Config) (*Plane, error) {
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if cfg.BaseURL == "" {
		return nil, errors.New("saasplane: BaseURL is required")
	}
	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("saasplane: invalid BaseURL: %w", err)
	}
	cfg.DaemonID = strings.TrimSpace(cfg.DaemonID)
	if cfg.DaemonID == "" {
		return nil, errors.New("saasplane: DaemonID is required")
	}
	cfg.DeviceID = strings.TrimSpace(cfg.DeviceID)
	if cfg.DeviceID == "" {
		cfg.DeviceID = cfg.DaemonID
	}
	cfg.DeviceSecret = strings.TrimSpace(cfg.DeviceSecret)
	cfg.BearerToken = strings.TrimSpace(cfg.BearerToken)
	if cfg.DeviceSecret != "" && cfg.DeviceID == "" {
		return nil, errors.New("saasplane: DeviceID is required when DeviceSecret is set")
	}
	cfg.Agents = normalizeAgents(cfg.Agents)
	if len(cfg.Agents) == 0 && cfg.DeviceSecret == "" {
		return nil, errors.New("saasplane: at least one static agent binding or a device credential is required")
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 5 * time.Second
	}
	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: cfg.RequestTimeout}
	}
	p := &Plane{
		cfg:    cfg,
		client: client,
		ops:    make(chan stateOp, 64),
		done:   make(chan struct{}),
	}
	go p.loop()
	return p, nil
}

func (p *Plane) Close() {
	ack := make(chan struct{})
	select {
	case p.ops <- stateOp{close: true, ack: ack}:
		<-ack
	case <-p.done:
	}
}

func (p *Plane) RegisterRuntime(ctx context.Context, rt controlplane.RuntimeRegistration) error {
	if !p.dynamicBindingsEnabled() {
		return nil
	}
	runtimeID := strings.TrimSpace(rt.RuntimeID)
	provider := providerFromRuntimeID(firstNonEmpty(rt.Provider, runtimeID))
	if runtimeID == "" || provider == "" {
		return nil
	}
	availability, detectionState := runtimeAvailability(rt, provider)
	var out struct {
		SchemaVersion string `json:"schema_version"`
	}
	return p.postJSON(ctx, "/v1/daemon/runtime-snapshot", DeviceRuntimeSnapshotSyncRequest{
		DaemonID:          firstNonEmpty(rt.DaemonID, p.cfg.DaemonID),
		DeviceID:          p.cfg.DeviceID,
		DeviceDisplayName: firstNonEmpty(rt.DeviceName, p.cfg.DeviceID),
		Runtimes: []RuntimeSnapshotRecord{{
			RuntimeID:                 runtimeID,
			Kind:                      runtimeKindForProvider(provider),
			Availability:              availability,
			DetectionState:            detectionState,
			RequiresExperimentalOptIn: runtimeRequiresExperimentalOptIn(rt, provider),
			Models:                    runtimeModels(rt.Models),
		}},
	}, &out)
}

func (p *Plane) DeregisterRuntime(context.Context, string) error {
	return nil
}

func runtimeAvailability(rt controlplane.RuntimeRegistration, provider string) (string, string) {
	if available, ok := rt.Capabilities["provider."+provider+".available"]; ok && !available {
		return "offline", "missing"
	}
	return "online", "detected"
}

func runtimeModels(in []controlplane.RuntimeModel) []RuntimeModelRecord {
	out := make([]RuntimeModelRecord, 0, len(in))
	for _, model := range in {
		out = append(out, RuntimeModelRecord{
			ModelID:   model.ModelID,
			Label:     model.Label,
			IsDefault: model.IsDefault,
		})
	}
	return out
}

func runtimeRequiresExperimentalOptIn(rt controlplane.RuntimeRegistration, provider string) bool {
	if len(rt.Capabilities) == 0 {
		return false
	}
	key := "provider." + provider + ".requires_experimental_opt_in"
	return rt.Capabilities[key]
}

func (p *Plane) Heartbeat(ctx context.Context, hb controlplane.RuntimeHeartbeat) error {
	if p.dynamicBindingsEnabled() {
		assignmentsByAgent, err := p.activeAssignmentsByAgentForHeartbeat(ctx, hb.RunningTaskIDs)
		if err != nil {
			return err
		}
		for agentID, assignmentIDs := range assignmentsByAgent {
			if len(assignmentIDs) == 0 {
				continue
			}
			var out assignmentcontract.AgentHeartbeatResponse
			if err := p.postJSON(ctx, "/v1/agents/"+url.PathEscape(agentID)+"/heartbeat", assignmentcontract.AgentHeartbeatRequest{
				DaemonID:            p.cfg.DaemonID,
				DeviceID:            p.cfg.DeviceID,
				RuntimeID:           hb.RuntimeID,
				RunningTaskIDs:      append([]string(nil), hb.RunningTaskIDs...),
				ActiveAssignmentIDs: assignmentIDs,
			}, &out); err != nil {
				return err
			}
		}
		return nil
	}
	agentID, ok := agentFromRuntimeID(hb.RuntimeID)
	if !ok {
		return nil
	}
	assignmentIDs, err := p.activeAssignmentIDsForHeartbeat(ctx, agentID, hb.RunningTaskIDs)
	if err != nil {
		return err
	}
	if len(assignmentIDs) == 0 {
		return nil
	}
	var out assignmentcontract.AgentHeartbeatResponse
	return p.postJSON(ctx, "/v1/agents/"+url.PathEscape(agentID)+"/heartbeat", assignmentcontract.AgentHeartbeatRequest{
		DaemonID:            p.cfg.DaemonID,
		DeviceID:            p.cfg.DeviceID,
		RuntimeID:           hb.RuntimeID,
		RunningTaskIDs:      append([]string(nil), hb.RunningTaskIDs...),
		ActiveAssignmentIDs: assignmentIDs,
	}, &out)
}

func (p *Plane) ClaimTask(ctx context.Context, runtimeID string) (*bridge.TaskRequest, error) {
	provider := providerFromRuntimeID(runtimeID)
	if p.dynamicBindingsEnabled() {
		bindings, err := p.agentBindings(ctx)
		if err != nil {
			return nil, err
		}
		for _, binding := range bindings {
			if binding.RuntimeProvider != provider || strings.TrimSpace(binding.RuntimeID) != strings.TrimSpace(runtimeID) {
				continue
			}
			poll, err := p.pollAgent(ctx, binding.AgentID, runtimeID)
			if err != nil {
				return nil, err
			}
			if poll.Assignment == nil {
				continue
			}
			switch poll.Action {
			case assignmentcontract.PollStart, assignmentcontract.PollActive:
				assignment := *poll.Assignment
				if assignment.RuntimeProvider != "" && assignment.RuntimeProvider != provider {
					continue
				}
				if err := p.saveAssignmentRuntime(ctx, assignment, runtimeID); err != nil {
					return nil, err
				}
				return taskRequestFromAssignment(assignment), nil
			case assignmentcontract.PollCancel:
				_ = p.deliverCancel(ctx, *poll.Assignment)
				return nil, nil
			case assignmentcontract.PollNone:
				continue
			default:
				continue
			}
		}
		return nil, nil
	}
	runtimeAgent, hasRuntimeAgent := agentFromRuntimeID(runtimeID)
	for _, agent := range p.cfg.Agents {
		if agent.RuntimeProvider != provider {
			continue
		}
		if hasRuntimeAgent && agent.AgentID != runtimeAgent {
			continue
		}
		poll, err := p.pollAgent(ctx, agent.AgentID, runtimeID)
		if err != nil {
			return nil, err
		}
		if poll.Assignment == nil {
			continue
		}
		switch poll.Action {
		case assignmentcontract.PollStart, assignmentcontract.PollActive:
			assignment := *poll.Assignment
			if assignment.RuntimeProvider != "" && assignment.RuntimeProvider != provider {
				continue
			}
			if err := p.saveAssignmentRuntime(ctx, assignment, runtimeID); err != nil {
				return nil, err
			}
			return taskRequestFromAssignment(assignment), nil
		case assignmentcontract.PollCancel:
			_ = p.deliverCancel(ctx, *poll.Assignment)
			return nil, nil
		case assignmentcontract.PollNone:
			continue
		default:
			continue
		}
	}
	return nil, nil
}

func (p *Plane) WatchCancellation(ctx context.Context, taskID string) (<-chan error, error) {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		return nil, errors.New("saasplane: empty taskID")
	}
	ch := make(chan error, 1)
	err := p.withState(ctx, func(s *planeState) {
		s.cancelWatchers[taskID] = ch
	})
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func (p *Plane) StartTask(ctx context.Context, taskID string) error {
	assignment, ok, err := p.assignmentForTask(ctx, taskID)
	if err != nil || !ok {
		return err
	}
	_, err = p.postAgentEvent(ctx, assignment, assignmentcontract.AgentEventRequest{
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
		State:        assignmentcontract.AssignmentReady,
		EventType:    assignmentcontract.EventAssignmentReady,
		Message:      "daemon ready",
	})
	return err
}

func (p *Plane) ReportEvent(ctx context.Context, taskID string, ev agentbridge.Event) error {
	assignment, ok, err := p.assignmentForTask(ctx, taskID)
	if err != nil || !ok {
		return err
	}
	req, ok := eventRequestFromAgentEvent(assignment, ev)
	if !ok {
		return nil
	}
	_, err = p.postAgentEvent(ctx, assignment, req)
	return err
}

func (p *Plane) CompleteTask(ctx context.Context, taskID string, res agentbridge.Result) error {
	assignment, ok, err := p.assignmentForTask(ctx, taskID)
	if err != nil || !ok {
		return err
	}
	state, eventType := terminalStateAndEvent(res.Status)
	message := res.Error
	if message == "" {
		message = res.Output
	}
	_, err = p.postAgentEvent(ctx, assignment, assignmentcontract.AgentEventRequest{
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
		State:        state,
		EventType:    eventType,
		Message:      message,
	})
	if err != nil {
		return err
	}
	return p.withState(ctx, func(s *planeState) {
		delete(s.assignmentsByTask, taskID)
		delete(s.runtimeIDsByTask, taskID)
		delete(s.cancelWatchers, taskID)
	})
}

func (p *Plane) pollAgent(ctx context.Context, agentID, runtimeID string) (assignmentcontract.PollResponse, error) {
	var out assignmentcontract.PollResponse
	err := p.postJSON(ctx, "/v1/agents/"+url.PathEscape(agentID)+"/poll", assignmentcontract.PollRequest{
		DaemonID:  p.cfg.DaemonID,
		DeviceID:  p.cfg.DeviceID,
		RuntimeID: runtimeID,
	}, &out)
	return out, err
}

func (p *Plane) postAgentEvent(ctx context.Context, assignment assignmentcontract.Assignment, req assignmentcontract.AgentEventRequest) (assignmentcontract.AgentEventResponse, error) {
	var out assignmentcontract.AgentEventResponse
	req.DaemonID = p.cfg.DaemonID
	req.DeviceID = p.cfg.DeviceID
	runtimeID, err := p.runtimeIDForAssignment(ctx, assignment)
	if err != nil {
		return out, err
	}
	req.RuntimeID = runtimeID
	err = p.postJSON(ctx, "/v1/agents/"+url.PathEscape(assignment.AgentID)+"/events", req, &out)
	return out, err
}

func (p *Plane) postJSON(ctx context.Context, path string, in any, out any) error {
	ctx, cancel := context.WithTimeout(ctx, p.cfg.RequestTimeout)
	defer cancel()
	body, err := json.Marshal(in)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.cfg.DeviceSecret != "" {
		req.Header.Set("X-Riido-Device-ID", p.cfg.DeviceID)
		req.Header.Set("X-Riido-Device-Secret", p.cfg.DeviceSecret)
	}
	if p.cfg.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.BearerToken)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("saasplane: %s returned %s: %s", path, resp.Status, strings.TrimSpace(string(b)))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (p *Plane) getJSON(ctx context.Context, path string, out any) error {
	ctx, cancel := context.WithTimeout(ctx, p.cfg.RequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.BaseURL+path, nil)
	if err != nil {
		return err
	}
	if p.cfg.DeviceSecret != "" {
		req.Header.Set("X-Riido-Device-ID", p.cfg.DeviceID)
		req.Header.Set("X-Riido-Device-Secret", p.cfg.DeviceSecret)
	}
	if p.cfg.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.cfg.BearerToken)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("saasplane: %s returned %s: %s", path, resp.Status, strings.TrimSpace(string(b)))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (p *Plane) saveAssignmentRuntime(ctx context.Context, assignment assignmentcontract.Assignment, runtimeID string) error {
	return p.withState(ctx, func(s *planeState) {
		s.assignmentsByTask[assignment.TaskID] = assignment
		if runtimeID != "" {
			s.runtimeIDsByTask[assignment.TaskID] = runtimeID
		}
	})
}

func (p *Plane) assignmentForTask(ctx context.Context, taskID string) (assignmentcontract.Assignment, bool, error) {
	var assignment assignmentcontract.Assignment
	var ok bool
	err := p.withState(ctx, func(s *planeState) {
		assignment, ok = s.assignmentsByTask[taskID]
	})
	return assignment, ok, err
}

func (p *Plane) agentBindings(ctx context.Context) ([]assignmentcontract.AgentRuntimeBinding, error) {
	var out AgentRuntimeBindingListResponse
	if err := p.getJSON(ctx, "/v1/daemon/agent-bindings", &out); err != nil {
		return nil, err
	}
	return out.Bindings, nil
}

func (p *Plane) activeAssignmentIDsForHeartbeat(ctx context.Context, agentID string, runningTaskIDs []string) ([]string, error) {
	var tasks []string
	seen := map[string]bool{}
	for _, taskID := range runningTaskIDs {
		taskID = strings.TrimSpace(taskID)
		if taskID != "" && !seen[taskID] {
			seen[taskID] = true
			tasks = append(tasks, taskID)
		}
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	var ids []string
	err := p.withState(ctx, func(s *planeState) {
		for _, taskID := range tasks {
			assignment := s.assignmentsByTask[taskID]
			if assignment.AgentID != agentID || assignment.ID == "" {
				continue
			}
			ids = append(ids, assignment.ID)
		}
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (p *Plane) activeAssignmentsByAgentForHeartbeat(ctx context.Context, runningTaskIDs []string) (map[string][]string, error) {
	tasks := normalizedTaskIDs(runningTaskIDs)
	if len(tasks) == 0 {
		return nil, nil
	}
	byAgent := map[string][]string{}
	err := p.withState(ctx, func(s *planeState) {
		for _, taskID := range tasks {
			assignment := s.assignmentsByTask[taskID]
			if assignment.AgentID == "" || assignment.ID == "" {
				continue
			}
			byAgent[assignment.AgentID] = append(byAgent[assignment.AgentID], assignment.ID)
		}
	})
	return byAgent, err
}

func (p *Plane) deliverCancel(ctx context.Context, assignment assignmentcontract.Assignment) error {
	return p.withState(ctx, func(s *planeState) {
		ch := s.cancelWatchers[assignment.TaskID]
		if ch == nil {
			return
		}
		select {
		case ch <- fmt.Errorf("saas assignment %s cancelled", assignment.ID):
		default:
		}
	})
}

type planeState struct {
	assignmentsByTask map[string]assignmentcontract.Assignment
	runtimeIDsByTask  map[string]string
	cancelWatchers    map[string]chan error
}

type stateOp struct {
	fn    func(*planeState)
	close bool
	ack   chan struct{}
}

func (p *Plane) loop() {
	state := planeState{
		assignmentsByTask: map[string]assignmentcontract.Assignment{},
		runtimeIDsByTask:  map[string]string{},
		cancelWatchers:    map[string]chan error{},
	}
	defer close(p.done)
	for op := range p.ops {
		if op.close {
			for _, ch := range state.cancelWatchers {
				close(ch)
			}
			close(op.ack)
			return
		}
		op.fn(&state)
		close(op.ack)
	}
}

func (p *Plane) withState(ctx context.Context, fn func(*planeState)) error {
	ack := make(chan struct{})
	select {
	case p.ops <- stateOp{fn: fn, ack: ack}:
	case <-p.done:
		return errors.New("saasplane: closed")
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case <-ack:
		return nil
	case <-p.done:
		return errors.New("saasplane: closed")
	case <-ctx.Done():
		return ctx.Err()
	}
}

func taskRequestFromAssignment(assignment assignmentcontract.Assignment) *bridge.TaskRequest {
	metadata := map[string]string{
		MetadataAssignmentID:    assignment.ID,
		MetadataAgentID:         assignment.AgentID,
		MetadataComponentID:     assignment.ComponentID,
		MetadataLeaseToken:      assignment.LeaseToken,
		MetadataModelID:         assignment.ModelID,
		MetadataRuntimeProvider: assignment.RuntimeProvider,
		"workspace_id":          firstNonEmpty(assignment.ComponentID, assignment.TaskID),
		"run_id":                assignment.ID,
	}
	prompt, systemPrompt, telemetryPlacement, instructionPlacement := agentbridge.ApplyRuntimeInstructionContract(assignment.RuntimeProvider, assignment.Prompt, "", assignment.AgentInstruction)
	metadata[agentbridge.MetadataTelemetryContract] = telemetryPlacement
	if instructionPlacement != "" {
		metadata[agentbridge.MetadataAgentInstruction] = instructionPlacement
	}
	return &bridge.TaskRequest{
		ID:                       assignment.TaskID,
		Provider:                 bridge.Provider(assignment.RuntimeProvider),
		Model:                    providerModelOverride(assignment.RuntimeProvider, assignment.ModelID),
		Prompt:                   prompt,
		SystemPrompt:             systemPrompt,
		AllowExperimentalRuntime: assignment.AllowExperimentalRuntime,
		Metadata:                 metadata,
	}
}

func providerModelOverride(runtimeProvider, modelID string) string {
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		return ""
	}
	switch strings.TrimSpace(runtimeProvider) {
	case "codex":
		if modelID == "codex-default" {
			return ""
		}
	case "claude", "claude_code":
		if modelID == "claude-default" {
			return ""
		}
	case "openclaw":
		if modelID == "openclaw-default" {
			return ""
		}
	case "cursor":
		if modelID == "cursor-auto" {
			return ""
		}
	}
	if modelID == "runtime-default" {
		return ""
	}
	return modelID
}

func eventRequestFromAgentEvent(assignment assignmentcontract.Assignment, ev agentbridge.Event) (assignmentcontract.AgentEventRequest, bool) {
	req := assignmentcontract.AgentEventRequest{
		AssignmentID: assignment.ID,
		TaskID:       assignment.TaskID,
	}
	switch ev.Kind {
	case agentbridge.EventProgress:
		req.EventType = assignmentcontract.EventRiidoLog
		req.Message = ev.Text
	case agentbridge.EventLifecycle:
		if ev.Phase == agentbridge.StateRunning {
			req.EventType = assignmentcontract.EventAssignmentRunning
			req.State = assignmentcontract.AssignmentRunning
			req.Message = "provider running"
		} else {
			return req, false
		}
	case agentbridge.EventLog:
		req.EventType = assignmentcontract.EventProviderLog
		req.Message = ev.Text
	case agentbridge.EventWarning:
		req.EventType = assignmentcontract.EventProviderWarning
		req.Message = ev.Text
	case agentbridge.EventError:
		req.EventType = assignmentcontract.EventProviderError
		req.Message = firstNonEmpty(ev.Err, ev.Text)
	default:
		return req, false
	}
	return req, true
}

func terminalStateAndEvent(status agentbridge.ResultStatus) (assignmentcontract.AssignmentState, string) {
	switch status {
	case agentbridge.ResultCompleted:
		return assignmentcontract.AssignmentCompleted, assignmentcontract.EventAssignmentCompleted
	case agentbridge.ResultCancelled:
		return assignmentcontract.AssignmentCancelled, assignmentcontract.EventAssignmentCancelled
	default:
		return assignmentcontract.AssignmentFailed, assignmentcontract.EventAssignmentFailed
	}
}

func providerFromRuntimeID(runtimeID string) string {
	parts := strings.Split(runtimeID, ":")
	return strings.TrimSpace(parts[len(parts)-1])
}

func RuntimeIDForAgent(daemonID string, agent AgentBinding) string {
	return strings.TrimSpace(daemonID) + ":agent:" + url.QueryEscape(strings.TrimSpace(agent.AgentID)) + ":" + strings.TrimSpace(agent.RuntimeProvider)
}

func (p *Plane) runtimeIDForAssignment(ctx context.Context, assignment assignmentcontract.Assignment) (string, error) {
	if p.dynamicBindingsEnabled() {
		var runtimeID string
		err := p.withState(ctx, func(s *planeState) {
			runtimeID = s.runtimeIDsByTask[assignment.TaskID]
		})
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(runtimeID) != "" {
			return runtimeID, nil
		}
	}
	for _, agent := range p.cfg.Agents {
		if agent.AgentID == assignment.AgentID && agent.RuntimeProvider == assignment.RuntimeProvider {
			return RuntimeIDForAgent(p.cfg.DaemonID, agent), nil
		}
	}
	return RuntimeIDForAgent(p.cfg.DaemonID, AgentBinding{AgentID: assignment.AgentID, RuntimeProvider: assignment.RuntimeProvider}), nil
}

func agentFromRuntimeID(runtimeID string) (string, bool) {
	parts := strings.Split(runtimeID, ":")
	if len(parts) < 4 || parts[len(parts)-3] != "agent" {
		return "", false
	}
	agentID, err := url.QueryUnescape(strings.TrimSpace(parts[len(parts)-2]))
	if err != nil {
		return "", false
	}
	agentID = strings.TrimSpace(agentID)
	return agentID, agentID != ""
}

func normalizeAgents(in []AgentBinding) []AgentBinding {
	out := make([]AgentBinding, 0, len(in))
	for _, agent := range in {
		agent.AgentID = strings.TrimSpace(agent.AgentID)
		agent.RuntimeProvider = strings.TrimSpace(agent.RuntimeProvider)
		if agent.AgentID == "" || agent.RuntimeProvider == "" {
			continue
		}
		out = append(out, agent)
	}
	return out
}

func (p *Plane) dynamicBindingsEnabled() bool {
	return len(p.cfg.Agents) == 0
}

func normalizedTaskIDs(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, taskID := range in {
		taskID = strings.TrimSpace(taskID)
		if taskID == "" || seen[taskID] {
			continue
		}
		seen[taskID] = true
		out = append(out, taskID)
	}
	return out
}

func runtimeKindForProvider(provider string) string {
	switch strings.TrimSpace(provider) {
	case "claude":
		return "claude_code"
	default:
		return strings.TrimSpace(provider)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
