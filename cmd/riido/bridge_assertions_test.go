package main

import (
	"slices"
	"testing"
)

func assertBridgeArgPair(t *testing.T, args []string, key, value string) {
	t.Helper()
	for i := 0; i+1 < len(args); i++ {
		if args[i] == key && args[i+1] == value {
			return
		}
	}
	t.Fatalf("missing arg pair %s %s in %v", key, value, args)
}

func containsEnv(env []string, want string) bool {
	return slices.Contains(env, want)
}
