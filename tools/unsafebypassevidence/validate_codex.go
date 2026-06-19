package main

import (
	"slices"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/provider/codex"
)

func validateCodexArgs(rows []Surface) ([]problem, []CodexArgEvidence) {
	var problems []problem
	var evidence []CodexArgEvidence
	args := codex.UnsafeBypassArgs()
	for _, row := range rows {
		if row.Provider != "Codex" {
			continue
		}
		arg := strings.TrimSuffix(row.Flag, "=true")
		ok := slices.Contains(args, arg)
		evidence = append(evidence, CodexArgEvidence{Arg: arg, OK: ok})
		if !ok {
			problems = append(problems, problem{Message: "codex unsafe arg missing: " + arg})
		}
	}
	return problems, evidence
}
