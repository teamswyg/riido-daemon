package contextcli

import (
	"testing"
	"time"
)

func TestNoSecretInArgv(t *testing.T) {
	cases := [][]string{
		{"--token", "abc"},
		{"--token=abc"},
		{"get", "RIID-1", "--api-key", "k"},
		{"--password=x"},
		{"--refresh-token", "r"},
	}
	for _, args := range cases {
		_, err := ParseArgs(args, ArgSpec{MaxPositionals: 2})
		if err == nil || err.Code != CodeSecretInArgv {
			t.Fatalf("args %v: want SECRET_IN_ARGV, got %v", args, err)
		}
		if err.ExitCode() != ExitUsage {
			t.Fatalf("args %v: want exit 2, got %d", args, err.ExitCode())
		}
	}
}

func TestParseArgsCommonAndSpecificFlags(t *testing.T) {
	args, err := ParseArgs(
		[]string{"RIID-5839", "--json", "--limit=50", "--timeout", "5", "--budget", "4000", "--check-api"},
		ArgSpec{Values: []string{"--budget"}, Bools: []string{"--check-api"}, MaxPositionals: 1},
	)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !args.JSON || args.Limit != 50 || args.Timeout != 5*time.Second {
		t.Fatalf("common options not parsed: %+v", args.Options)
	}
	if args.Value("--budget") != "4000" || !args.Bool("--check-api") {
		t.Fatalf("specific flags not parsed: budget=%q check=%v", args.Value("--budget"), args.Bool("--check-api"))
	}
	if len(args.Positionals) != 1 || args.Positionals[0] != "RIID-5839" {
		t.Fatalf("positionals: %v", args.Positionals)
	}
}

func TestParseArgsDefaultsAndRejections(t *testing.T) {
	args, err := ParseArgs(nil, ArgSpec{})
	if err != nil || args.Limit != DefaultLimit || args.Timeout != DefaultTimeout || args.JSON {
		t.Fatalf("defaults: %+v err=%v", args.Options, err)
	}
	bad := [][]string{
		{"--limit", "0"},
		{"--limit", "101"},
		{"--timeout", "0s"},
		{"--format", "yaml"},
		{"--unknown"},
		{"--json=true"},
		{"a", "b"},
		{"--workspace"},
	}
	for _, in := range bad {
		if _, err := ParseArgs(in, ArgSpec{MaxPositionals: 1}); err == nil || err.Code != CodeUsage {
			t.Fatalf("args %v: want USAGE, got %v", in, err)
		}
	}
}
