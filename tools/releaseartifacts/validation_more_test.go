package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateManifestReportsRequiredCollections(t *testing.T) {
	m := manifest{
		SchemaVersion:    "bad",
		ReleaseWorkflow:  "",
		CDNPublishScript: "",
		DetailDocs:       []detailDoc{{Title: "Only title"}},
		Targets:          []target{{Platform: "darwin"}},
	}
	problems := validateManifest(m)
	for _, want := range []string{
		"unexpected schema_version",
		"id, title, generated_doc, and workflow are required",
		"release workflow and scripts are required",
		"CDN publish workflow and script are required",
		"targets and four detail docs are required",
		"detail docs require title and path",
		"archive content and forbidden item rules are required",
		"invalid target",
		"installer command and CDN base URL are required",
	} {
		assertReleaseProblem(t, problems, want)
	}
}

func TestSourceAndDocChecksReportEdges(t *testing.T) {
	dir := t.TempDir()
	mustReleaseWrite(t, filepath.Join(dir, "script.sh"), "anchor")
	checks := []sourceCheck{
		{Name: "ok", File: "script.sh", Contains: "anchor"},
		{Name: "missing", File: "script.sh", Contains: "absent"},
		{Name: "nofile", File: "missing.sh", Contains: "anchor"},
	}
	results := checkSources(dir, checks)
	if len(failedChecks("source check failed", results)) != 2 || !results[0].Pass {
		t.Fatalf("unexpected source results %#v", results)
	}
	docs := map[string]string{"docs/release.md": "fresh"}
	assertReleaseProblem(t, checkDocs(options{Repo: dir, CheckDoc: true}, docs), "docs/release.md")
	if err := maybeWriteDocs(options{Repo: dir, WriteDoc: true}, docs); err != nil {
		t.Fatal(err)
	}
	if got := checkDocs(options{Repo: dir, CheckDoc: true}, docs); len(got) != 0 {
		t.Fatalf("fresh doc returned %#v", got)
	}
	mustReleaseWrite(t, filepath.Join(dir, "docs/release.md"), "stale")
	assertReleaseProblem(t, checkDocs(options{Repo: dir, CheckDoc: true}, docs), "generated doc drift")
}

func assertReleaseProblem(t *testing.T, problems []problem, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem.Message, needle) {
			return
		}
	}
	t.Fatalf("missing %q in %#v", needle, problems)
}
