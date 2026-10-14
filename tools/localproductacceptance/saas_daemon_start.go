package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func startSaaSDaemon(binary string, slot saasDaemonSlot, host string) scenario {
	sc := scenario{ID: saasStartScenarioID(slot.Index), Method: "DAEMON", Endpoint: slot.Socket}
	if err := os.MkdirAll(filepath.Dir(slot.Socket), 0o755); err != nil {
		sc.Status = statusFailed
		sc.FailureSummary = err.Error()
		return sc
	}
	out, err := startSaaSDaemonCommand(binary, slot, host)
	sc.Observed = map[string]any{
		"device_id":   slot.Credential.DeviceID,
		"socket":      slot.Socket,
		"pid_file":    slot.PIDFile,
		"output_tail": outputTail(out),
	}
	if err != nil {
		sc.Status = statusFailed
		sc.FailureSummary = err.Error()
		return sc
	}
	sc.Status = statusPassed
	return sc
}

func saasStartScenarioID(slot int) string {
	return "local.saas.daemon_start." + strconv.Itoa(slot)
}

func startSaaSDaemonCommand(binary string, slot saasDaemonSlot, host string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "daemon", "start",
		"--socket", slot.Socket, "--pid-file", slot.PIDFile,
		"--log-file", slot.LogFile, "--lock-file", slot.LockFile)
	cmd.Env = append(os.Environ(), saasDaemonEnv(slot, host)...)
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		return string(out), ctx.Err()
	}
	return string(out), err
}
