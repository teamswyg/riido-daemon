package main

import (
	"strings"
	"testing"
)

func assertRepoError(t *testing.T, err error, needle string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), needle) {
		t.Fatalf("expected %q error, got %v", needle, err)
	}
}

func assertRepoProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem, needle) {
			return
		}
	}
	t.Fatalf("missing %q in %#v", needle, problems)
}
