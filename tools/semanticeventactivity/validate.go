package main

import "fmt"

const schemaVersion = "riido-semantic-event-activity.v1"

func validate(repo string, manifest Manifest) []problem {
	var problems []problem
	if manifest.SchemaVersion != schemaVersion {
		problems = append(problems, problem{fmt.Sprintf("schema_version must be %s", schemaVersion)})
	}
	required := []string{manifest.ID, manifest.Title, manifest.GeneratedDoc, manifest.Workflow, manifest.EvidenceArtifact}
	for _, value := range required {
		if value == "" {
			problems = append(problems, problem{"manifest has empty required string"})
		}
	}
	return append(problems, validateClassifications(repo, manifest)...)
}

func validateClassifications(repo string, manifest Manifest) []problem {
	manifestKinds, problems := manifestKindMap(manifest)
	sourceKinds, err := sourceEventKinds(repo)
	if err != nil {
		problems = append(problems, problem{err.Error()})
	}
	for kind := range sourceKinds {
		if _, ok := runtimeKinds()[kind]; !ok {
			problems = append(problems, problem{"runtime catalog missing source event kind: " + kind})
		}
	}
	for kind := range runtimeKinds() {
		if _, ok := sourceKinds[kind]; !ok {
			problems = append(problems, problem{"runtime catalog declares unknown source event kind: " + kind})
		}
	}
	for kind, semantic := range runtimeKinds() {
		got, ok := manifestKinds[kind]
		if !ok {
			problems = append(problems, problem{"manifest missing runtime event kind: " + kind})
			continue
		}
		if got != semantic {
			problems = append(problems, problem{fmt.Sprintf("%s category drift: manifest=%s runtime=%s", kind, category(got), category(semantic))})
		}
	}
	for kind := range manifestKinds {
		if _, ok := runtimeKinds()[kind]; !ok {
			problems = append(problems, problem{"manifest declares unknown event kind: " + kind})
		}
	}
	return problems
}
