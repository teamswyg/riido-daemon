package detectutil

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func defaultLoginShellPATH() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	shell := loginShell()
	if shell == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), loginShellPATHTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, shell, "-lc", loginShellPATHScript())
	cmd.Stdin = nil
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return ""
	}
	return extractLoginShellPATH(buf.String())
}

func loginShell() string {
	return strings.TrimSpace(os.Getenv("SHELL"))
}

func loginShellPATHScript() string {
	return "printf '%s' \"" + loginPATHMarkerStart + "${PATH}" + loginPATHMarkerEnd + "\""
}

func extractLoginShellPATH(out string) string {
	start := strings.Index(out, loginPATHMarkerStart)
	end := strings.Index(out, loginPATHMarkerEnd)
	if start < 0 || end < 0 || end < start {
		return ""
	}
	return out[start+len(loginPATHMarkerStart) : end]
}
