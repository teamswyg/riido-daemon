// Package riidoapi exposes Riido's local-only daemon API.
//
// The API intentionally uses a tiny local JSON envelope: one local IPC request
// with a method and optional params, one JSON response envelope.
// It is the first surface that GUI/Zed integrations can consume without
// reading Riido's state files directly.
package riidoapi

import (
	"os"
	"path/filepath"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/validation"
)

const (
	StatusSchemaVersion     = "riido-api-status.v1"
	ReviewDemoSchemaVersion = "riido-api-review-demo.v1"
)

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
