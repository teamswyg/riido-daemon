package main

func buildEvidence(
	manifest Manifest,
	fsm FSMSnapshot,
	problems []problem,
	sources []SourceCheckEvidence,
	forbiddenOK bool,
) Evidence {
	return Evidence{
		ID:             manifest.ID,
		SchemaVersion:  manifest.SchemaVersion,
		GeneratedDoc:   manifest.GeneratedDoc,
		Workflow:       manifest.Workflow,
		SourcePackage:  manifest.SourcePackage,
		Problems:       problems,
		FSM:            fsm,
		SourceChecks:   sources,
		ForbiddenCheck: ForbiddenCheck{Tokens: manifest.ForbiddenDocTokens, OK: forbiddenOK},
	}
}
