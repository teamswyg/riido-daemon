package riidoapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

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

func (c Client) Request(ctx context.Context, method string, params, out any) error {
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
