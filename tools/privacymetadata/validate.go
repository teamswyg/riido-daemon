package main

func validate(repo string, manifest Manifest, policy PolicySnapshot) (
	[]problem,
	[]SourceCheckEvidence,
	[]ShapeCheck,
) {
	sourceProblems, sources := validateSources(repo, manifest.SourceChecks)
	shapeProblems, shapes := validateShapes(policy)
	sourceProblems = append(sourceProblems, shapeProblems...)
	return sourceProblems, sources, shapes
}
