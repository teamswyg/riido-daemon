package contextcli

import "strings"

// Environment variables the daemon injects into agent runtime processes.
const (
	EnvRunToken     = "RIIDO_RUN_TOKEN"
	EnvDaemonSocket = "RIIDO_DAEMON_SOCKET"
	EnvTaskID       = "RIIDO_TASK_ID"
	EnvWorkspaceID  = "RIIDO_WORKSPACE_ID"
)

// Mode selects how a work-context command reaches Riido.
type Mode string

const (
	// ModeAgent calls the local daemon broker with the run capability.
	ModeAgent Mode = "agent"
	// ModeUser calls Riido servers with the signed-in user's credential.
	ModeUser Mode = "user"
)

// Environment is the mode-selection input read from process environment.
//
// Mode selection is not a security boundary: a shell-capable agent can unset
// RIIDO_RUN_TOKEN. Server permission checks and OS user-presence protection of
// the user credential are the boundary.
type Environment struct {
	Mode         Mode   `json:"mode"`
	DaemonSocket string `json:"daemon_socket,omitempty"`
	TaskID       string `json:"task_id,omitempty"`
	WorkspaceID  string `json:"workspace_id,omitempty"`
	runToken     string
}

// DetectEnvironment reads the mode-selection variables through getenv.
func DetectEnvironment(getenv func(string) string) Environment {
	env := Environment{
		Mode:         ModeUser,
		DaemonSocket: strings.TrimSpace(getenv(EnvDaemonSocket)),
		TaskID:       strings.TrimSpace(getenv(EnvTaskID)),
		WorkspaceID:  strings.TrimSpace(getenv(EnvWorkspaceID)),
		runToken:     strings.TrimSpace(getenv(EnvRunToken)),
	}
	if env.runToken != "" {
		env.Mode = ModeAgent
	}
	return env
}

// RunCapability returns the run capability for the daemon broker. It must
// only be sent over local IPC and never printed or logged.
func (e Environment) RunCapability() string {
	return e.runToken
}

// HasRunCapability reports whether a run capability is present.
func (e Environment) HasRunCapability() bool {
	return e.runToken != ""
}
