package toolpolicy

import "testing"

func TestProtectedConfigPathMatchesCredentialFiles(t *testing.T) {
	for _, path := range []string{
		".docker/config.json",
		"repo/.docker/config.json",
		".config/gh/hosts.yml",
		"repo/.config/gh/hosts.yml",
	} {
		if !isProtectedConfigPath(path) {
			t.Fatalf("path should be protected: %q", path)
		}
	}
}

func TestProtectedConfigPathRejectsNeighborFiles(t *testing.T) {
	for _, path := range []string{
		".docker/daemon.json",
		".config/gh/config.yml",
		"docs/.docker/config.json.example",
		"docs/.config/gh/hosts.yml.example",
	} {
		if isProtectedConfigPath(path) {
			t.Fatalf("path should not be protected config: %q", path)
		}
	}
}

func TestNormalizePathCanonicalizesUserInput(t *testing.T) {
	got := normalizePath(` "./Repo\\.Docker\\Config.JSON" `)
	if got != "repo/.docker/config.json" {
		t.Fatalf("normalized path=%q", got)
	}
}

func TestProtectedPathAllowsClaudeExtensionDirs(t *testing.T) {
	for _, path := range []string{
		".claude/commands/deploy.md",
		".claude/agents/reviewer.md",
		".claude/skills/custom/SKILL.md",
		".claude/worktrees/demo/config.json",
	} {
		if isProtectedPath(path) {
			t.Fatalf("claude extension path should be allowed: %q", path)
		}
	}
}
