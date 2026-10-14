package main

func validateFigmaSections(rows []figmaSection) []string {
	var problems []string
	if len(rows) == 0 {
		return []string{"figma sections must not be empty"}
	}
	for _, row := range rows {
		if len(row.Refs) == 0 || row.DaemonScope == "" || len(row.NotOwned) == 0 {
			problems = append(problems, "figma sections require refs, daemon_scope, and not_owned facts")
		}
	}
	return problems
}
