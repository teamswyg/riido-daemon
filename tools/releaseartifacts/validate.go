package main

import "fmt"

func validateManifest(m manifest) []problem {
	var problems []problem
	if m.SchemaVersion != "riido-release-artifacts.v1" {
		problems = append(problems, problem{Message: "unexpected schema_version"})
	}
	if m.ID == "" || m.Title == "" || m.GeneratedDoc == "" || m.Workflow == "" {
		problems = append(problems, problem{Message: "id, title, generated_doc, and workflow are required"})
	}
	if m.ReleaseWorkflow == "" || m.BuildScript == "" || m.PublishScript == "" || m.InstallScript == "" {
		problems = append(problems, problem{Message: "release workflow and scripts are required"})
	}
	if m.CDNPublishWorkflow == "" || m.CDNPublishScript == "" {
		problems = append(problems, problem{Message: "CDN publish workflow and script are required"})
	}
	return append(problems, validateCollections(m)...)
}

func validateCollections(m manifest) []problem {
	var problems []problem
	if len(m.Targets) == 0 || len(m.DetailDocs) != 4 {
		problems = append(problems, problem{Message: "targets and four detail docs are required"})
	}
	for _, doc := range m.DetailDocs {
		if doc.Title == "" || doc.Path == "" {
			problems = append(problems, problem{Message: "detail docs require title and path"})
		}
	}
	if len(m.ArchiveContents) == 0 || len(m.ForbiddenArchiveItems) == 0 {
		problems = append(problems, problem{Message: "archive content and forbidden item rules are required"})
	}
	for _, t := range m.Targets {
		if t.Platform == "" || t.GOOS == "" || t.GOARCH == "" || t.Format == "" {
			problems = append(problems, problem{Message: fmt.Sprintf("invalid target: %+v", t)})
		}
	}
	if m.Installer.Command == "" || m.DesktopMSIX.CDNLatestBaseURL == "" {
		problems = append(problems, problem{Message: "installer command and CDN base URL are required"})
	}
	return problems
}
