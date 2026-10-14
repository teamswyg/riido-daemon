package main

func validate(repo string, manifest Manifest) ([]problem, []InterfaceCheck, []SourceResult) {
	interfaceProblems, interfaceChecks := validateInterfaces(repo, manifest)
	sourceProblems, sourceChecks := validateSources(repo, manifest)
	problems := interfaceProblems
	problems = append(problems, sourceProblems...)
	return problems, interfaceChecks, sourceChecks
}
