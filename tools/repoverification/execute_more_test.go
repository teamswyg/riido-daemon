package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunCommandCapturesPassFailAndOutputTail(t *testing.T) {
	helper := os.Args[0]
	t.Setenv("RIIDO_REPOVERIFICATION_HELPER", "1")
	pass := runCommand(t.TempDir(), commandSpec{
		ID: "pass", Argv: []string{helper, "-test.run=TestRepoVerificationHelperProcess", "--", "pass"},
	})
	if pass.Status != "passed" || pass.OutputTail != "ok" {
		t.Fatalf("unexpected passed evidence: %#v", pass)
	}
	fail := runCommand(t.TempDir(), commandSpec{
		ID: "fail", Argv: []string{helper, "-test.run=TestRepoVerificationHelperProcess", "--", "fail"},
	})
	if fail.Status != "failed" || !strings.Contains(fail.OutputTail, "boom") {
		t.Fatalf("unexpected failed evidence: %#v", fail)
	}
	long := runCommand(t.TempDir(), commandSpec{
		ID: "long", Argv: []string{helper, "-test.run=TestRepoVerificationHelperProcess", "--", "long"},
	})
	if len(long.OutputTail) > 700 || !strings.Contains(long.OutputTail, "tail") {
		t.Fatalf("unexpected compacted output: len=%d tail=%q", len(long.OutputTail), long.OutputTail)
	}
}

func TestRunCommandsAndCommandValidationEdges(t *testing.T) {
	helper := os.Args[0]
	t.Setenv("RIIDO_REPOVERIFICATION_HELPER", "1")
	specs := []commandSpec{
		{ID: "a", Argv: []string{helper, "-test.run=TestRepoVerificationHelperProcess", "--", "pass"}},
		{ID: "b", Argv: []string{helper, "-test.run=TestRepoVerificationHelperProcess", "--", "fail"}},
	}
	out := runCommands(t.TempDir(), specs)
	if len(out) != 2 || !anyFailed(out) {
		t.Fatalf("expected one failed command: %#v", out)
	}
	problems := validateCommand(commandSpec{ID: "bad", Description: "bad", Argv: []string{""}}, map[string]bool{})
	if !strings.Contains(strings.Join(problems, "\n"), "argv[0]") {
		t.Fatalf("expected argv[0] validation problem: %#v", problems)
	}
	quoted := shellQuote([]string{"echo", "a b", "it's"})
	if quoted != "echo 'a b' 'it'\\''s'" {
		t.Fatalf("unexpected shell quote: %q", quoted)
	}
	if compactOutput(strings.Repeat("x", 701)) != strings.Repeat("x", 700) {
		t.Fatal("compact output should keep the last 700 bytes")
	}
}

func TestRepoVerificationHelperProcess(t *testing.T) {
	if os.Getenv("RIIDO_REPOVERIFICATION_HELPER") != "1" {
		return
	}
	switch os.Args[len(os.Args)-1] {
	case "pass":
		os.Stdout.WriteString("ok")
	case "fail":
		os.Stdout.WriteString("boom")
		os.Exit(3)
	case "long":
		os.Stdout.WriteString(strings.Repeat("x", 800) + "tail")
	default:
		os.Exit(2)
	}
	os.Exit(0)
}
