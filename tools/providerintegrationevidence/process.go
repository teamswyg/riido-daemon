package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	providerVersionTimeout     = 10 * time.Second
	providerIntegrationTimeout = 5 * time.Minute
)

func resolveExecutable(defaultName, override string) (string, bool) {
	if override != "" {
		if info, err := os.Stat(override); err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return override, true
		}
	}
	path, err := exec.LookPath(defaultName)
	return path, err == nil
}

func probeVersion(exe string) string {
	ctx, cancel := context.WithTimeout(context.Background(), providerVersionTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, exe, "--version").CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func runIntegrationTest(root string, provider provider) (string, string) {
	ctx, cancel := context.WithTimeout(context.Background(), providerIntegrationTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", provider.GoPackage, "-race", "-run", provider.TestRegex, "-v")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "AGENTBRIDGE_INTEGRATION=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "failed", fmt.Sprintf("timed out after %s", providerIntegrationTimeout)
		}
		return "failed", compactOutput(string(out))
	}
	return "passed", ""
}

func integrationCommand(provider provider) string {
	return "AGENTBRIDGE_INTEGRATION=1 go test " + provider.GoPackage + " -race -run " + provider.TestRegex + " -v"
}
