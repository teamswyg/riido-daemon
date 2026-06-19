package main

func validate(repo string, manifest Manifest) (
	[]problem,
	[]SourceCheckEvidence,
	[]AbsentEvidence,
) {
	var problems []problem
	problems = append(problems, validateRefs(manifest)...)
	sourceProblems, sources := validateSources(repo, manifest.SourceChecks)
	absentProblems, absent := validateAbsent(repo, manifest.AbsentSurfaces)
	problems = append(problems, sourceProblems...)
	problems = append(problems, absentProblems...)
	return problems, sources, absent
}
