package validation

import "time"

const (
	DefaultGate    = "deterministic-command-exit-code.v1"
	DefaultTimeout = 5 * time.Minute
)

type CommandRequest struct {
	Command        string
	Workdir        string
	Timeout        time.Duration
	CommandID      string
	Provider       string
	ValidationGate string
	Summary        string
}

type CommandResult struct {
	Command           string `json:"command"`
	Workdir           string `json:"workdir"`
	ExitCode          int    `json:"exit_code"`
	Result            string `json:"result"`
	ValidationGate    string `json:"validation_gate"`
	ProviderRunID     string `json:"provider_run_id"`
	ProviderRunResult string `json:"provider_run_result"`
	Summary           string `json:"summary"`
	StartedAt         string `json:"started_at"`
	FinishedAt        string `json:"finished_at"`
}
