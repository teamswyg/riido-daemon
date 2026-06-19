package main

func validate(repo string, manifest Manifest) (
	[]problem,
	[]SourceResult,
	[]AbsentCheck,
) {
	sourceProblems, sources := validateSources(repo, manifest.SourceChecks)
	absentProblems, absent := validateAbsent(repo, manifest.AbsentSurfaces)
	problems := sourceProblems
	problems = append(problems, absentProblems...)
	problems = append(problems, validateStepReferences(manifest)...)
	return problems, sources, absent
}
