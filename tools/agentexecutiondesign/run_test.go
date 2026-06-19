package main

import (
	"strings"
	"testing"
)

func TestRenderedDocsUseEvidenceFiles(t *testing.T) {
	m, err := loadModel("../..", "docs/30-architecture/agent-execution-unresolved-design.riido.json")
	if err != nil {
		t.Fatal(err)
	}
	docs := renderedDocs(m)
	body := docs[baseDir+"verification-evidence.md"]
	for _, want := range []string{
		"same-task-multiple-assignments",
		"TestRuntimeActorUsesAssignmentIDAsExecutionKey",
		"private-repo-url-redaction",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("verification doc missing %q\n%s", want, body)
		}
	}
}
