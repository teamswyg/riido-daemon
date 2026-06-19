package main

import (
	"fmt"
	"os"
	"strings"
)

const manifestSchema = "riido-roadmap-open-questions.v1"

func validateManifest(repo string, m manifest) ([]string, []sourceCheckResult) {
	var problems []string
	if m.SchemaVersion != manifestSchema {
		problems = append(problems, "unexpected schema_version")
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" {
		problems = append(problems, "id, title, and generated_doc are required")
	}
	if m.Workflow == "" || m.EvidenceArtifact == "" {
		problems = append(problems, "workflow and evidence_artifact are required")
	}
	if len(m.Questions) == 0 || len(m.Assertions) == 0 {
		problems = append(problems, "questions and assertions are required")
	}
	problems = append(problems, validateQuestions(m.Questions)...)
	results, sourceProblems := validateSourceChecks(repo, m.SourceChecks)
	problems = append(problems, sourceProblems...)
	problems = append(problems, mustExist(repo, m.Workflow)...)
	return problems, results
}

func validateQuestions(questions []question) []string {
	var problems []string
	seen := map[string]bool{}
	for _, q := range questions {
		if q.ID == "" || q.Area == "" || q.Question == "" || q.CurrentHandling == "" {
			problems = append(problems, "question fields are required")
		}
		if !strings.HasPrefix(q.ID, "Q-") {
			problems = append(problems, "question id must start with Q-: "+q.ID)
		}
		if seen[q.ID] {
			problems = append(problems, "duplicate question id "+q.ID)
		}
		seen[q.ID] = true
	}
	return problems
}

func mustExist(repo, rel string) []string {
	if _, err := os.Stat(repoPath(repo, rel)); err != nil {
		return []string{fmt.Sprintf("missing artifact %q", rel)}
	}
	return nil
}
