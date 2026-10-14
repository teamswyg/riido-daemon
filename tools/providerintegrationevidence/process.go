package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

const (
	providerVersionTimeout     = 10 * time.Second
	providerIntegrationTimeout = 5 * time.Minute
)

func resolveExecutable(defaultName, override string) (string, bool) {
	return detectutil.ResolveExecutable(defaultName, override)
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
	cmd := exec.CommandContext(ctx, "go", "test", provider.GoPackage, "-race", "-count=1", "-run", provider.TestRegex, "-v")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "AGENTBRIDGE_INTEGRATION=1")
	out, err := cmd.CombinedOutput()
	summary := compactOutput(string(out))
	if integrationSkipped(string(out)) {
		return "skipped", summary
	}
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "failed", fmt.Sprintf("timed out after %s", providerIntegrationTimeout)
		}
		return "failed", summary
	}
	return "passed", ""
}

func integrationCommand(provider provider) string {
	return "AGENTBRIDGE_INTEGRATION=1 go test " + provider.GoPackage + " -race -count=1 -run " + provider.TestRegex + " -v"
}

func integrationSkipped(out string) bool {
	return strings.Contains(out, "--- SKIP:")
}
