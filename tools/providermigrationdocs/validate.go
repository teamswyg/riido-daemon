package main

import "fmt"

const schemaVersion = "riido-provider-public-migration-docs.v1"

func validateManifest(repo string, m manifest) []string {
	var problems []string
	if m.SchemaVersion != schemaVersion {
		problems = append(problems, "unexpected schema_version")
	}
	if m.ID == "" || m.GeneratedDoc == "" || m.Workflow == "" || m.EvidenceArtifact == "" {
		problems = append(problems, "id, generated_doc, workflow, and evidence_artifact are required")
	}
	if len(m.Pages) == 0 || len(m.Assertions) == 0 {
		problems = append(problems, "pages and assertions must not be empty")
	}
	problems = append(problems, mustExist(repo, m.Workflow)...)
	problems = append(problems, validatePages(repo, m.Pages)...)
	return problems
}

func validatePages(repo string, pages []page) []string {
	var problems []string
	ids := map[string]bool{}
	providers := map[string]bool{}
	for _, page := range pages {
		problems = append(problems, validatePage(repo, page, ids, providers)...)
	}
	return problems
}

func validatePage(repo string, page page, ids, providers map[string]bool) []string {
	var problems []string
	if page.ID == "" || page.Title == "" || page.GeneratedDoc == "" {
		problems = append(problems, "page id, title, and generated_doc are required")
	}
	if ids[page.ID] {
		problems = append(problems, "duplicate page id "+page.ID)
	}
	ids[page.ID] = true
	if page.ProviderID != "" && providers[page.ProviderID] {
		problems = append(problems, "duplicate provider id "+page.ProviderID)
	}
	providers[page.ProviderID] = page.ProviderID != ""
	problems = append(problems, validateTable(page)...)
	for _, artifact := range page.Artifacts {
		problems = append(problems, mustExist(repo, artifact)...)
	}
	return problems
}

func validateTable(page page) []string {
	if len(page.TableRows) == 0 {
		return nil
	}
	var problems []string
	for _, row := range page.TableRows {
		if len(row) != len(page.TableColumns) {
			problems = append(problems, fmt.Sprintf("table row width mismatch on %s", page.ID))
		}
	}
	return problems
}
