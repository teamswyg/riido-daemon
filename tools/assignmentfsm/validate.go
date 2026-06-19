package main

func validate(repo string, manifest Manifest, rendered string) (
	[]problem,
	[]SourceCheckEvidence,
	bool,
) {
	sourceProblems, sources := validateSources(repo, manifest.SourceChecks)
	forbiddenProblems := validateForbiddenDocTokens(rendered, manifest.ForbiddenDocTokens)
	sourceProblems = append(sourceProblems, forbiddenProblems...)
	return sourceProblems, sources, len(forbiddenProblems) == 0
}
