package main

import (
	"os/exec"
	"path/filepath"
	"strings"
)

func gitChangedFiles(repo string) []string {
	if out, ok := gitOutput(repo, "diff", "--name-only", "--cached"); ok {
		files := splitLines(out)
		if len(files) > 0 {
			return files
		}
	}
	if out, ok := gitOutput(repo, "diff", "--name-only", "HEAD"); ok {
		files := splitLines(out)
		files = append(files, untrackedFiles(repo)...)
		return dedupe(files)
	}
	return nil
}

func gitOutput(repo string, args ...string) (string, bool) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo
	out, err := cmd.Output()
	if err == nil {
		return string(out), true
	}
	return gitOutputWithWorkTree(repo, args...)
}

func gitOutputWithWorkTree(repo string, args ...string) (string, bool) {
	gitDir := filepath.Join(repo, ".git")
	all := append([]string{"--git-dir=" + gitDir, "--work-tree=" + repo}, args...)
	cmd := exec.Command("git", all...)
	out, err := cmd.Output()
	return string(out), err == nil
}

func splitLines(s string) []string {
	var out []string
	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func untrackedFiles(repo string) []string {
	out, ok := gitOutput(repo, "ls-files", "--others", "--exclude-standard")
	if !ok {
		return nil
	}
	return splitLines(out)
}

func dedupe(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
