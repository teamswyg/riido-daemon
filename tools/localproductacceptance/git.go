package main

import (
	"os/exec"
	"strings"
)

func clientReadOnlyScenario(root string) scenario {
	dirty, dirtyDetail := gitOutput(root, "status", "--porcelain", "--untracked-files=no")
	tracked, trackedDetail := gitOutput(root, "ls-files", "e2e/ai-agent")
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
		Summary: "Product QA harness code must live outside teamswyg/riido-client; close or abandon client changes.",
	}
}
