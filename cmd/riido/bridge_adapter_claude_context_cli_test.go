package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestClaudeContextCLIAllowsOnlySupportedReadLeaves(t *testing.T) {
	req := agentbridge.StartRequest{Env: map[string]string{
		"RIIDO_RUN_TOKEN": "cap-secret", "RIIDO_DAEMON_SOCKET": "/tmp/agentd.sock", "RIIDO_TASK_ID": "task-1",
	}, ContextCLIEnabled: true, CustomArgs: []string{"--allowedTools", "Bash(*)", "--allowed-tools=Bash(riido *)"}}
	cmd, err := (bridgeClaudeAdapter{}).BuildStart(req)
	if err != nil {
		t.Fatal(err)
	}
	idx := slices.Index(cmd.Args, "--allowedTools")
	if idx < 0 || idx+1 >= len(cmd.Args) {
		t.Fatal("missing native read allowlist")
	}
	rules := strings.Split(cmd.Args[idx+1], ",")
	if len(rules) != 18 {
		t.Fatalf("read rules = %v", rules)
	}
	for _, leaf := range []string{"context get", "work get", "work tree", "work relations", "work comments", "agents list", "agents show", "meeting list", "meeting get"} {
		for _, suffix := range []string{"", " *"} {
			if !slices.Contains(rules, "Bash(riido "+leaf+suffix+")") {
				t.Errorf("missing read rule %s%s", leaf, suffix)
			}
		}
	}
	args := strings.Join(cmd.Args, " ")
	for _, forbidden := range []string{"cap-secret", "Bash(*)", "Bash(riido *)", "Bash(riido meeting *)", "Bash(riido meeting update *)", "bypassPermissions"} {
		if strings.Contains(args, forbidden) {
			t.Fatalf("unexpected argument %q", forbidden)
		}
	}
}

func TestClaudeContextCLIIncompleteWiringKeepsPriorPermissionArgs(t *testing.T) {
	for _, env := range []map[string]string{nil, {"RIIDO_RUN_TOKEN": "cap"}, {"RIIDO_DAEMON_SOCKET": "/tmp/sock", "RIIDO_TASK_ID": "task-1"}} {
		cmd, err := (bridgeClaudeAdapter{}).BuildStart(agentbridge.StartRequest{Env: env, ContextCLIEnabled: true})
		if err != nil {
			t.Fatal(err)
		}
		if slices.Contains(cmd.Args, "--allowedTools") {
			t.Fatal("disabled/incomplete CLI changed permissions")
		}
	}
	cmd, err := (bridgeClaudeAdapter{}).BuildStart(agentbridge.StartRequest{Env: map[string]string{
		"RIIDO_RUN_TOKEN": "cap", "RIIDO_DAEMON_SOCKET": "/tmp/sock", "RIIDO_TASK_ID": "task-1",
	}})
	if err != nil || slices.Contains(cmd.Args, "--allowedTools") {
		t.Fatal("environment alone enabled CLI allowlist without a launch hook")
	}
}
