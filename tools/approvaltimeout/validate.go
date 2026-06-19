package main

func validate(repo string, manifest Manifest) ([]problem, []ManifestCheck, []SourceResult) {
	semantic, err := loadJSON[SemanticActivityManifest](repo, manifest.Sources.SemanticActivityManifest)
	if err != nil {
		return []problem{{err.Error()}}, nil, nil
	}
	draft, err := loadJSON[ProviderDraftManifest](repo, manifest.Sources.ProviderDraftManifest)
	if err != nil {
		return []problem{{err.Error()}}, nil, nil
	}
	manifestProblems, manifestChecks := validateManifests(manifest, semantic, draft)
	sourceProblems, sourceChecks := validateSources(repo, manifest)
	problems := manifestProblems
	problems = append(problems, sourceProblems...)
	return problems, manifestChecks, sourceChecks
}
