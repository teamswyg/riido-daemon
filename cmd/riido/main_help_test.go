package main

import "testing"

func TestTopLevelHelpReturnsNil(t *testing.T) {
	if err := run([]string{"--help"}); err != nil {
		t.Fatalf("help returned error: %v", err)
	}
}
