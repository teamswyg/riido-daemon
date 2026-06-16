package main

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

const daemonPIDIdentitySchemaVersion = "riido-daemon-pid-identity.v1"

type daemonPIDIdentity struct {
	SchemaVersion string    `json:"schema_version"`
	PID           int       `json:"pid"`
	Executable    string    `json:"executable,omitempty"`
	Socket        string    `json:"socket"`
	StartedAt     time.Time `json:"started_at"`
}

func writeDaemonPIDFiles(pidFile, socket string) error {
	pid := os.Getpid()
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0o644); err != nil {
		return err
	}
	executable, _ := os.Executable()
	identity := daemonPIDIdentity{
		SchemaVersion: daemonPIDIdentitySchemaVersion,
		PID:           pid,
		Executable:    executable,
		Socket:        socket,
		StartedAt:     time.Now().UTC(),
	}
	body, err := json.MarshalIndent(identity, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(daemonPIDIdentityPath(pidFile), append(body, '\n'), 0o600)
}

func removeDaemonPIDFiles(pidFile string) {
	_ = os.Remove(pidFile)
	_ = os.Remove(daemonPIDIdentityPath(pidFile))
}

func loadDaemonPIDIdentity(pidFile string) (daemonPIDIdentity, bool, error) {
	raw, err := os.ReadFile(daemonPIDIdentityPath(pidFile))
	if errors.Is(err, os.ErrNotExist) {
		return daemonPIDIdentity{}, false, nil
	}
	if err != nil {
		return daemonPIDIdentity{}, false, err
	}
	var identity daemonPIDIdentity
	if err := json.Unmarshal(raw, &identity); err != nil {
		return daemonPIDIdentity{}, true, err
	}
	return identity, true, nil
}

func daemonPIDIdentityPath(pidFile string) string {
	return strings.TrimSpace(pidFile) + ".identity.json"
}
