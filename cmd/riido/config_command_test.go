package main

import "testing"

func TestConfigRootHelp(t *testing.T) {
	if err := run([]string{"config", "--help"}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceUseHelp(t *testing.T) {
	h := newContextHarness(t, nil)
	if err := runWorkspace(h.io, []string{"use", "--help"}); err != nil {
		t.Fatal(err)
	}
}
