package agentexecutionevidence

import "path/filepath"

var (
	repoRoot     = filepath.Join("..", "..")
	manifestPath = filepath.Join(
		repoRoot,
		"docs",
		"30-architecture",
		"agent-execution-unresolved-design",
		"assignment-lifecycle-evidence.riido.json",
	)
	humanDocPath = filepath.Join(
		repoRoot,
		"docs",
		"30-architecture",
		"agent-execution-unresolved-design",
		"assignment-lifecycle-fsm.md",
	)
)
