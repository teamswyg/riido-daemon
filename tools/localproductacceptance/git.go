package main

import (
	"os/exec"
	"strings"
)

const clientBoundarySummary = "Product QA harness code and Codex work must never be merged into teamswyg/riido-client."

func clientReadOnlyScenario(root string) scenario {
	dirty, dirtyDetail := gitOutput(root, "status", "--porcelain", "--untracked-files=no")
	tracked, trackedDetail := trackedHarnessFiles(root)
	if dirty != "" || tracked != "" {
		return scenario{
			ID:             "product.client.readonly",
			Status:         statusFailed,
			FailureSummary: strings.TrimSpace(dirtyDetail + "\n" + trackedDetail),
			Repair:         clientReadOnlyRepair(),
		}
	}
	return scenario{ID: "product.client.readonly", Status: statusPassed}
}

func trackedHarnessFiles(root string) (string, string) {
	return gitOutput(root, "ls-files", "e2e/ai-agent", ".riido-local")
}

func gitOutput(root string, args ...string) (string, string) {
	allArgs := append([]string{"-C", root}, args...)
	out, err := exec.Command("git", allArgs...).CombinedOutput()
	text := strings.TrimSpace(string(out))
	if err != nil {
		return text, "git " + strings.Join(args, " ") + ": " + text
	}
	return text, text
}

func clientReadOnlyRepair() *repair {
	return &repair{
		Class:   "client_repo_must_remain_read_only",
		Owner:   "local-qa",
		Mode:    "manual",
		Summary: clientBoundarySummary + " Keep experiments in riido-daemon and close or abandon client changes.",
	}
}
