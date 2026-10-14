package main

import "testing"

func TestHasExecutableStepDetectsShellRun(t *testing.T) {
	workflow := `jobs:
  check:
    steps:
      - name: Verify shell gate
        run: scripts/verify-riido-work-branch.sh "$GITHUB_HEAD_REF"
`
	if !hasExecutableStep(workflow) {
		t.Fatal("shell run step was not detected as executable")
	}
}
