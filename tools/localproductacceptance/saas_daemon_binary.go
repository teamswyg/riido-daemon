package main

import (
	"os"
	"path/filepath"
	"time"
)

func buildLocalDaemonBinary(path string) scenario {
	sc := scenario{ID: "local.saas.daemon_binary.build", Method: "GO", Endpoint: path}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		sc.Status = statusFailed
		sc.FailureSummary = err.Error()
		return sc
	}
	out, err := runLocalCommand(2*time.Minute, "go", "build", "-o", path, "./cmd/riido")
	sc.Observed = map[string]any{"output_tail": outputTail(out)}
	if err != nil {
		sc.Status = statusFailed
		sc.FailureSummary = err.Error()
		return sc
	}
	sc.Status = statusPassed
	return sc
}
