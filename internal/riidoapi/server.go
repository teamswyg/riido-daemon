// Package riidoapi exposes Riido's local-only daemon API.
//
// The API intentionally uses a tiny local JSON envelope: one local IPC request
// with a method and optional params, one JSON response envelope.
// It is the first surface that GUI/Zed integrations can consume without
// reading Riido's state files directly.
package riidoapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/validation"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

const StatusSchemaVersion = "riido-api-status.v1"
const ReviewDemoSchemaVersion = "riido-api-review-demo.v1"

type LocalTransport string

const (
	LocalTransportUnixSocket       LocalTransport = "unix-socket"
	LocalTransportWindowsNamedPipe LocalTransport = "windows-named-pipe"
)

type Config struct {
	SocketPath string         `json:"socket_path"`
	TaskDBPath string         `json:"task_db_path"`
	Transport  LocalTransport `json:"transport"`
}

type Server struct {
	config Config
}

type Status struct {
	SchemaVersion       string `json:"schema_version"`
	Transport           string `json:"transport"`
	SocketPath          string `json:"socket_path"`
	TaskDBPath          string `json:"task_db_path"`
	TaskDBSchemaVersion string `json:"task_db_schema_version"`
	TaskCount           int    `json:"task_count"`
	TransitionCount     int    `json:"transition_count"`
	EvidenceCount       int    `json:"evidence_count"`
	CommandReceiptCount int    `json:"command_receipt_count"`
	DiagnosticCount     int    `json:"diagnostic_count"`
	UpdatedAt           string `json:"updated_at"`
}

type TransitionRequest struct {
	TaskID      string `json:"task_id"`
	ToState     string `json:"to_state"`
	EventType   string `json:"event_type"`
	Actor       string `json:"actor"`
	Source      string `json:"source"`
	Reason      string `json:"reason"`
	Provider    string `json:"provider"`
	DecisionLLM string `json:"decision_llm"`
	ApprovalID  string `json:"approval_id"`
	CommandID   string `json:"command_id"`
}

type TransitionResponse struct {
	TaskDBPath string                          `json:"task_db_path"`
	Task       taskdb.TaskRecord               `json:"task"`
	Transition taskdb.TaskTransitionRecord     `json:"transition"`
	Receipt    taskdb.TaskCommandReceiptRecord `json:"receipt"`
}

type EvidenceRequest struct {
	TaskID            string `json:"task_id"`
	Command           string `json:"command"`
	ExitCode          int    `json:"exit_code"`
	Result            string `json:"result"`
	Actor             string `json:"actor"`
	Source            string `json:"source"`
	Summary           string `json:"summary"`
	Provider          string `json:"provider"`
	DecisionLLM       string `json:"decision_llm"`
	ApprovalID        string `json:"approval_id"`
	CommandID         string `json:"command_id"`
	ValidationGate    string `json:"validation_gate"`
	ProviderRunID     string `json:"provider_run_id"`
	ProviderRunResult string `json:"provider_run_result"`
}

type EvidenceResponse struct {
	TaskDBPath string                          `json:"task_db_path"`
	Task       taskdb.TaskRecord               `json:"task"`
	Evidence   taskdb.TaskEvidenceRecord       `json:"evidence"`
	Receipt    taskdb.TaskCommandReceiptRecord `json:"receipt"`
}

type ValidateRequest struct {
	TaskID         string `json:"task_id"`
	Command        string `json:"command"`
	Workdir        string `json:"workdir"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Actor          string `json:"actor"`
	Source         string `json:"source"`
	Summary        string `json:"summary"`
	Provider       string `json:"provider"`
	DecisionLLM    string `json:"decision_llm"`
	ApprovalID     string `json:"approval_id"`
	CommandID      string `json:"command_id"`
	ValidationGate string `json:"validation_gate"`
}

type ValidateResponse struct {
	TaskDBPath        string                           `json:"task_db_path"`
	Task              taskdb.TaskRecord                `json:"task"`
	Validation        validation.CommandResult         `json:"validation"`
	Evidence          taskdb.TaskEvidenceRecord        `json:"evidence"`
	Receipt           taskdb.TaskCommandReceiptRecord  `json:"receipt"`
	Transition        *taskdb.TaskTransitionRecord     `json:"transition,omitempty"`
	TransitionReceipt *taskdb.TaskCommandReceiptRecord `json:"transition_receipt,omitempty"`
}

type ReviewDemoRequest struct {
	DistributionChannel      string `json:"distribution_channel"`
	ReviewDemoConsentGranted bool   `json:"review_demo_consent_granted"`
}

type ReviewDemoResponse struct {
	SchemaVersion            string   `json:"schema_version"`
	DistributionChannel      string   `json:"distribution_channel"`
	Enabled                  bool     `json:"enabled"`
	Surfaces                 []string `json:"surfaces"`
	ProviderStatusMode       string   `json:"provider_status_mode"`
	ProviderExecutionAllowed bool     `json:"provider_execution_allowed"`
	TelemetrySyncAllowed     bool     `json:"telemetry_sync_allowed"`
	LocalOnly                bool     `json:"local_only"`
}

type Client struct {
	SocketPath string
	Transport  LocalTransport
	Timeout    time.Duration
}

func DefaultSocketPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "riido", "riido.sock"), nil
}

func NewServer(config Config) Server {
	return Server{config: config}
}

func NewClient(socketPath string) Client {
	return NewClientWithTransport(LocalTransportUnixSocket, socketPath)
}

func NewClientWithTransport(transport LocalTransport, socketPath string) Client {
	return Client{
		SocketPath: socketPath,
		Transport:  normalizeLocalTransport(transport),
		Timeout:    3 * time.Second,
	}
}

func (s Server) Serve(ctx context.Context) error {
	transport := normalizeLocalTransport(s.config.Transport)
	if s.config.SocketPath == "" {
		return errors.New("riido API socket path is empty")
	}
	if s.config.TaskDBPath == "" {
		return errors.New("riido task DB path is empty")
	}
	listener, cleanup, err := listenLocalEndpoint(transport, s.config.SocketPath)
	if err != nil {
		return fmt.Errorf("listen riido API %s endpoint: %w", transport, err)
	}
	defer func() {
		cleanup()
	}()

	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("accept riido API connection: %w", err)
		}
		go s.handleConn(conn)
	}
}

func (s Server) handleConn(conn net.Conn) {
	defer conn.Close()
	var req requestEnvelope
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		_ = writeResponse(conn, responseEnvelope{OK: false, Error: fmt.Sprintf("decode request: %v", err)})
		return
	}
	response := s.handleRequest(req)
	_ = writeResponse(conn, response)
}

func (s Server) handleRequest(req requestEnvelope) responseEnvelope {
	switch req.Method {
	case "status":
		db, err := taskdb.LoadTaskDBOrEmpty(s.config.TaskDBPath)
		if err != nil {
			return errorResponse(req.Method, err)
		}
		return okResponse(req.Method, statusFromDB(s.config, db))
	case "tasks":
		db, err := taskdb.LoadTaskDBOrEmpty(s.config.TaskDBPath)
		if err != nil {
			return errorResponse(req.Method, err)
		}
		return okResponse(req.Method, db)
	case "transition":
		response, err := s.applyTransition(req.Params)
		if err != nil {
			return errorResponse(req.Method, err)
		}
		return okResponse(req.Method, response)
	case "evidence":
		response, err := s.addEvidence(req.Params)
		if err != nil {
			return errorResponse(req.Method, err)
		}
		return okResponse(req.Method, response)
	case "validate":
		response, err := s.validateTask(req.Params)
		if err != nil {
			return errorResponse(req.Method, err)
		}
		return okResponse(req.Method, response)
	case "review-demo":
		response, err := s.evaluateReviewDemo(req.Params)
		if err != nil {
			return errorResponse(req.Method, err)
		}
		return okResponse(req.Method, response)
	default:
		return errorResponse(req.Method, fmt.Errorf("unknown method: %s", req.Method))
	}
}

func (s Server) evaluateReviewDemo(params json.RawMessage) (ReviewDemoResponse, error) {
	var req ReviewDemoRequest
	if len(params) == 0 {
		return ReviewDemoResponse{}, errors.New("review-demo params are required")
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return ReviewDemoResponse{}, fmt.Errorf("decode review-demo params: %w", err)
	}
	channel := hostintegration.DistributionChannel(strings.TrimSpace(req.DistributionChannel))
	if !channel.Valid() {
		return ReviewDemoResponse{}, fmt.Errorf("unknown distribution channel %q", req.DistributionChannel)
	}
	mode, err := hostintegration.EvaluateReviewDemoMode(hostintegration.ReviewDemoModeInput{
		Channel: channel,
		Consent: hostintegration.ConsentState{
			ReviewDemoMode: req.ReviewDemoConsentGranted,
		},
	})
	if err != nil {
		return ReviewDemoResponse{}, err
	}
	return reviewDemoResponseFromMode(mode), nil
}

func (s Server) validateTask(params json.RawMessage) (ValidateResponse, error) {
	var req ValidateRequest
	if len(params) == 0 {
		return ValidateResponse{}, errors.New("validate params are required")
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return ValidateResponse{}, fmt.Errorf("decode validate params: %w", err)
	}
	taskID := strings.TrimSpace(req.TaskID)
	if taskID == "" {
		return ValidateResponse{}, errors.New("task_id is required")
	}
	if strings.TrimSpace(req.Command) == "" {
		return ValidateResponse{}, errors.New("command is required")
	}
	if strings.TrimSpace(req.ApprovalID) == "" {
		return ValidateResponse{}, errors.New("approval_id is required before validation command execution")
	}
	if req.TimeoutSeconds < 0 {
		return ValidateResponse{}, errors.New("timeout_seconds must not be negative")
	}
	db, err := taskdb.LoadTaskDB(s.config.TaskDBPath)
	if err != nil {
		return ValidateResponse{}, err
	}
	providerForRun, err := validationProviderForTask(db, taskID, req.Provider)
	if err != nil {
		return ValidateResponse{}, err
	}
	if err := validateDecisionLLMForTask(db, taskID, req.DecisionLLM); err != nil {
		return ValidateResponse{}, err
	}
	taskBeforeValidation, ok := findTask(db, taskID)
	if !ok {
		return ValidateResponse{}, fmt.Errorf("task %s not found", taskID)
	}

	now := time.Now()
	commandID := strings.TrimSpace(req.CommandID)
	if commandID == "" {
		commandID = validationCommandID(taskID, now)
	}
	timeout := time.Duration(req.TimeoutSeconds) * time.Second
	actor := textutil.Default(req.Actor, "daemon")
	source := textutil.Default(req.Source, "riido-api")
	result, err := validation.RunCommand(context.Background(), validation.CommandRequest{
		Command:        req.Command,
		Workdir:        req.Workdir,
		Timeout:        timeout,
		CommandID:      commandID,
		Provider:       providerForRun,
		ValidationGate: req.ValidationGate,
		Summary:        req.Summary,
	}, now)
	if err != nil {
		return ValidateResponse{}, err
	}
	updated, evidence, receipt, err := taskdb.AddGuardedTaskEvidence(db, taskdb.TaskEvidenceInput{
		TaskID:            taskID,
		Command:           result.Command,
		ExitCode:          result.ExitCode,
		Result:            result.Result,
		Actor:             actor,
		Source:            source,
		Summary:           result.Summary,
		ValidationGate:    result.ValidationGate,
		ProviderRunID:     result.ProviderRunID,
		ProviderRunResult: result.ProviderRunResult,
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   commandID,
			Provider:    providerForRun,
			DecisionLLM: req.DecisionLLM,
			ApprovalID:  req.ApprovalID,
		},
	}, now)
	if err != nil {
		return ValidateResponse{}, err
	}

	var transition *taskdb.TaskTransitionRecord
	var transitionReceipt *taskdb.TaskCommandReceiptRecord
	if taskBeforeValidation.State == task.StateValidating {
		toState, eventType := validationTransitionForResult(result.Result)
		nextDB, appliedTransition, appliedReceipt, err := taskdb.ApplyGuardedTaskTransition(updated, taskdb.TaskTransitionInput{
			TaskID:  taskID,
			ToState: toState,
			Event:   eventType,
			Actor:   actor,
			Source:  source,
			Reason:  fmt.Sprintf("validation %s via %s", result.Result, result.ValidationGate),
			Guard: taskdb.TaskMutationGuardInput{
				CommandID:   commandID + ":transition",
				Provider:    providerForRun,
				DecisionLLM: req.DecisionLLM,
				ApprovalID:  req.ApprovalID,
			},
		}, now)
		if err != nil {
			return ValidateResponse{}, err
		}
		updated = nextDB
		transition = &appliedTransition
		transitionReceipt = &appliedReceipt
	}
	if err := taskdb.SaveTaskDB(s.config.TaskDBPath, updated); err != nil {
		return ValidateResponse{}, err
	}
	record, ok := findTask(updated, taskID)
	if !ok {
		return ValidateResponse{}, fmt.Errorf("task %s not found after validation", taskID)
	}
	return ValidateResponse{
		TaskDBPath:        s.config.TaskDBPath,
		Task:              record,
		Validation:        result,
		Evidence:          evidence,
		Receipt:           receipt,
		Transition:        transition,
		TransitionReceipt: transitionReceipt,
	}, nil
}

func (s Server) addEvidence(params json.RawMessage) (EvidenceResponse, error) {
	var req EvidenceRequest
	if len(params) == 0 {
		return EvidenceResponse{}, errors.New("evidence params are required")
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return EvidenceResponse{}, fmt.Errorf("decode evidence params: %w", err)
	}
	db, err := taskdb.LoadTaskDB(s.config.TaskDBPath)
	if err != nil {
		return EvidenceResponse{}, err
	}
	updated, evidence, receipt, err := taskdb.AddGuardedTaskEvidence(db, taskdb.TaskEvidenceInput{
		TaskID:            req.TaskID,
		Command:           req.Command,
		ExitCode:          req.ExitCode,
		Result:            req.Result,
		Actor:             req.Actor,
		Source:            req.Source,
		Summary:           req.Summary,
		ValidationGate:    req.ValidationGate,
		ProviderRunID:     req.ProviderRunID,
		ProviderRunResult: req.ProviderRunResult,
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   req.CommandID,
			Provider:    req.Provider,
			DecisionLLM: req.DecisionLLM,
			ApprovalID:  req.ApprovalID,
		},
	}, time.Now())
	if err != nil {
		return EvidenceResponse{}, err
	}
	if err := taskdb.SaveTaskDB(s.config.TaskDBPath, updated); err != nil {
		return EvidenceResponse{}, err
	}
	record, ok := findTask(updated, req.TaskID)
	if !ok {
		return EvidenceResponse{}, fmt.Errorf("task %s not found after evidence append", req.TaskID)
	}
	return EvidenceResponse{
		TaskDBPath: s.config.TaskDBPath,
		Task:       record,
		Evidence:   evidence,
		Receipt:    receipt,
	}, nil
}

func (s Server) applyTransition(params json.RawMessage) (TransitionResponse, error) {
	var req TransitionRequest
	if len(params) == 0 {
		return TransitionResponse{}, errors.New("transition params are required")
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return TransitionResponse{}, fmt.Errorf("decode transition params: %w", err)
	}
	to, err := taskdb.ParseTaskState(req.ToState)
	if err != nil {
		return TransitionResponse{}, err
	}
	db, err := taskdb.LoadTaskDB(s.config.TaskDBPath)
	if err != nil {
		return TransitionResponse{}, err
	}
	updated, transition, receipt, err := taskdb.ApplyGuardedTaskTransition(db, taskdb.TaskTransitionInput{
		TaskID:  req.TaskID,
		ToState: to,
		Event:   ir.EventType(req.EventType),
		Actor:   req.Actor,
		Source:  req.Source,
		Reason:  req.Reason,
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   req.CommandID,
			Provider:    req.Provider,
			DecisionLLM: req.DecisionLLM,
			ApprovalID:  req.ApprovalID,
		},
	}, time.Now())
	if err != nil {
		return TransitionResponse{}, err
	}
	if err := taskdb.SaveTaskDB(s.config.TaskDBPath, updated); err != nil {
		return TransitionResponse{}, err
	}
	record, ok := findTask(updated, req.TaskID)
	if !ok {
		return TransitionResponse{}, fmt.Errorf("task %s not found after transition", req.TaskID)
	}
	return TransitionResponse{
		TaskDBPath: s.config.TaskDBPath,
		Task:       record,
		Transition: transition,
		Receipt:    receipt,
	}, nil
}

func (c Client) Request(ctx context.Context, method string, params any, out any) error {
	transport := normalizeLocalTransport(c.Transport)
	if c.SocketPath == "" {
		return errors.New("riido API socket path is empty")
	}
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := dialLocalEndpoint(ctx, transport, c.SocketPath)
	if err != nil {
		return fmt.Errorf("connect riido API %s endpoint: %w", transport, err)
	}
	defer conn.Close()

	rawParams, err := rawParams(params)
	if err != nil {
		return err
	}
	requestBody, err := json.Marshal(requestEnvelope{Method: method, Params: rawParams})
	if err != nil {
		return fmt.Errorf("encode riido API request: %w", err)
	}
	if _, err := conn.Write(requestBody); err != nil {
		return fmt.Errorf("write riido API request: %w", err)
	}
	if unix, ok := conn.(*net.UnixConn); ok {
		if err := unix.CloseWrite(); err != nil {
			return fmt.Errorf("close riido API request stream: %w", err)
		}
	}

	responseBody, err := io.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("read riido API response: %w", err)
	}
	var env responseEnvelope
	if err := json.Unmarshal(responseBody, &env); err != nil {
		return fmt.Errorf("decode riido API response: %w", err)
	}
	if !env.OK {
		if env.Error != "" {
			return fmt.Errorf("riido API %s failed: %s", method, env.Error)
		}
		return fmt.Errorf("riido API %s failed", method)
	}
	if env.Method != method {
		return fmt.Errorf("riido API method mismatch: requested %s got %s", method, env.Method)
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("decode riido API %s data: %w", method, err)
	}
	return nil
}

func statusFromDB(config Config, db taskdb.TaskDB) Status {
	return Status{
		SchemaVersion:       StatusSchemaVersion,
		Transport:           string(normalizeLocalTransport(config.Transport)),
		SocketPath:          config.SocketPath,
		TaskDBPath:          config.TaskDBPath,
		TaskDBSchemaVersion: db.SchemaVersion,
		TaskCount:           len(db.Tasks),
		TransitionCount:     len(db.Transitions),
		EvidenceCount:       len(db.Evidence),
		CommandReceiptCount: len(db.CommandReceipts),
		DiagnosticCount:     len(db.Diagnostics),
		UpdatedAt:           db.UpdatedAt,
	}
}

func reviewDemoResponseFromMode(mode hostintegration.ReviewDemoMode) ReviewDemoResponse {
	surfaces := make([]string, 0, len(mode.Surfaces))
	for _, surface := range mode.Surfaces {
		surfaces = append(surfaces, string(surface))
	}
	providerStatusMode := "real-status"
	if mode.Enabled {
		providerStatusMode = "synthetic-preview"
	}
	return ReviewDemoResponse{
		SchemaVersion:            ReviewDemoSchemaVersion,
		DistributionChannel:      string(mode.Channel),
		Enabled:                  mode.Enabled,
		Surfaces:                 surfaces,
		ProviderStatusMode:       providerStatusMode,
		ProviderExecutionAllowed: mode.ProviderExecutionAllowed,
		TelemetrySyncAllowed:     mode.TelemetrySyncAllowed,
		LocalOnly:                true,
	}
}

func findTask(db taskdb.TaskDB, id string) (taskdb.TaskRecord, bool) {
	for _, record := range db.Tasks {
		if record.ID == id {
			return record, true
		}
	}
	return taskdb.TaskRecord{}, false
}

func validationProviderForTask(db taskdb.TaskDB, taskID string, requested string) (string, error) {
	taskRecord, ok := findTask(db, taskID)
	if !ok {
		return "", fmt.Errorf("task %s not found", taskID)
	}
	provider := strings.TrimSpace(requested)
	if provider == "" {
		provider = taskRecord.RecommendedProvider
	}
	if provider == "" {
		provider = db.RecommendedProvider
	}
	if provider == "" {
		return "", fmt.Errorf("task %s has no validation provider", taskID)
	}
	if !providerAvailable(db.ProviderCandidates, provider) {
		return "", fmt.Errorf("provider %s is not an available orchestration candidate for task %s", provider, taskID)
	}
	return provider, nil
}

func validateDecisionLLMForTask(db taskdb.TaskDB, taskID string, requested string) error {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return nil
	}
	taskRecord, ok := findTask(db, taskID)
	if !ok {
		return fmt.Errorf("task %s not found", taskID)
	}
	recommended := taskRecord.RecommendedDecisionLLM
	if recommended == "" {
		recommended = db.RecommendedDecisionLLM
	}
	if recommended != "" && requested != recommended {
		return fmt.Errorf("decision LLM %s does not match recommended decision LLM %s for task %s", requested, recommended, taskID)
	}
	return nil
}

func providerAvailable(candidates []taskdb.ProviderCandidate, provider string) bool {
	if len(candidates) == 0 {
		return true
	}
	for _, candidate := range candidates {
		if candidate.ID == provider {
			return candidate.Available
		}
	}
	return false
}

func validationCommandID(taskID string, now time.Time) string {
	return fmt.Sprintf("command:validation:%s:%s", taskID, now.UTC().Format("20060102T150405.000000000Z"))
}

func validationTransitionForResult(result string) (task.TaskState, ir.EventType) {
	if result == "passed" {
		return task.StatePatchReady, ir.EventValidationPassed
	}
	return task.StateFailed, ir.EventValidationFailed
}

func rawParams(params any) (json.RawMessage, error) {
	if params == nil {
		return nil, nil
	}
	data, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("encode riido API params: %w", err)
	}
	return data, nil
}

func okResponse(method string, data any) responseEnvelope {
	payload, err := json.Marshal(data)
	if err != nil {
		return errorResponse(method, err)
	}
	return responseEnvelope{
		OK:     true,
		Method: method,
		Data:   payload,
	}
}

func errorResponse(method string, err error) responseEnvelope {
	return responseEnvelope{
		OK:     false,
		Method: method,
		Error:  err.Error(),
	}
}

func writeResponse(conn net.Conn, response responseEnvelope) error {
	encoder := json.NewEncoder(conn)
	return encoder.Encode(response)
}

type requestEnvelope struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

type responseEnvelope struct {
	OK     bool            `json:"ok"`
	Method string          `json:"method"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
}
