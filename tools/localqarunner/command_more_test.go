package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunStepCapturesPassFailExitCodeAndTail(t *testing.T) {
	helper := os.Args[0]
	t.Setenv("RIIDO_LOCALQA_HELPER", "1")
	pass := runStep(t.TempDir(), "pass", helper, "-test.run=TestLocalQARunnerHelperProcess", "--", "pass")
	if pass.Status != statusPassed || pass.OutputTail != "ok" || pass.ExitCode != 0 {
		t.Fatalf("unexpected pass step: %#v", pass)
	}
	fail := runStep(t.TempDir(), "fail", helper, "-test.run=TestLocalQARunnerHelperProcess", "--", "fail")
	if fail.Status != statusFailed || fail.ExitCode != 7 || !strings.Contains(fail.OutputTail, "boom") {
		t.Fatalf("unexpected fail step: %#v", fail)
	}
	long := runStep(t.TempDir(), "long", helper, "-test.run=TestLocalQARunnerHelperProcess", "--", "long")
	if len(long.OutputTail) > 4000 || !strings.Contains(long.OutputTail, "tail") {
		t.Fatalf("unexpected long step tail: len=%d tail=%q", len(long.OutputTail), long.OutputTail)
	}
}

func TestEnvBoolAndExitHelpers(t *testing.T) {
	t.Setenv("RIIDO_LOCALQA_ENV_TEST", "")
	if getenvDefault("RIIDO_LOCALQA_ENV_TEST", "fallback") != "fallback" {
		t.Fatal("empty env should use fallback")
	}
	t.Setenv("RIIDO_LOCALQA_ENV_TEST", "value")
	if getenvDefault("RIIDO_LOCALQA_ENV_TEST", "fallback") != "value" {
		t.Fatal("non-empty env should win")
	}
	value := true
	if !boolValue(&value) || boolValue(nil) {
		t.Fatal("boolValue should be nil-safe")
	}
	if exitCode(os.ErrPermission) != -1 {
		t.Fatal("non-exit errors should map to -1")
	}
	env := strings.Join(localQAEnv(), "\n")
	if !strings.Contains(env, "PATH=") {
		t.Fatal("local QA env should include PATH")
	}
}

func TestLocalQARunnerHelperProcess(t *testing.T) {
	if os.Getenv("RIIDO_LOCALQA_HELPER") != "1" {
		return
	}
	switch os.Args[len(os.Args)-1] {
	case "pass":
		os.Stdout.WriteString("ok")
	case "fail":
		os.Stdout.WriteString("boom")
		os.Exit(7)
	case "long":
		os.Stdout.WriteString(strings.Repeat("x", 5000) + "tail")
	default:
		os.Exit(2)
	}
	os.Exit(0)
}
