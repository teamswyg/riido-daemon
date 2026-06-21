package main

import "testing"

func TestWorkflowEvidenceRequiresUploadedEvidenceOutPath(t *testing.T) {
	text := "" +
		"steps:\n" +
		"  - run: go run ./tools/example -evidence-out out/example.json\n" +
		"  - uses: actions/upload-artifact@v4\n" +
		"    with:\n" +
		"      name: example\n" +
		"      path: out/other.json\n" +
		"      if-no-files-found: error\n"
	record := workflowRecord{
		Path:                 ".github/workflows/example.yml",
		HasExecutable:        true,
		HasEvidenceOut:       true,
		EvidenceOutCount:     len(evidenceOutPaths(text)),
		UploadedEvidenceOut:  countUploadedEvidenceOut(evidenceOutPaths(text), artifactUploadPathValues(text)),
		MissingEvidenceOut:   missingEvidenceUploads(evidenceOutPaths(text), artifactUploadPathValues(text)),
		UploadsArtifact:      true,
		ArtifactUploadCount:  1,
		StrictUploadCount:    1,
		NonStrictUploadCount: 0,
	}
	got := classify(record, nil, nil)
	if got.Status != "missing_evidence_upload" {
		t.Fatalf("status = %q, missing = %#v", got.Status, got.MissingEvidenceOut)
	}
}

func TestVariableEvidenceOutBindsConcreteUploadPath(t *testing.T) {
	evidence := []string{"out/${boundary}-evidence.json"}
	uploads := []string{"out/runtime-snapshot-evidence.json"}
	if countUploadedEvidenceOut(evidence, uploads) != 1 {
		t.Fatalf("variable evidence-out did not bind to concrete upload")
	}
	if missing := missingEvidenceUploads(evidence, uploads); len(missing) != 0 {
		t.Fatalf("missing = %#v", missing)
	}
}

func TestGlobEvidenceUploadBindsEvidenceOut(t *testing.T) {
	evidence := []string{"out/distribution-host-docs.json"}
	uploads := []string{"out/*.json"}
	if countUploadedEvidenceOut(evidence, uploads) != 1 {
		t.Fatalf("glob upload did not bind evidence-out")
	}
}

func TestDirectoryEvidenceUploadBindsEvidenceOut(t *testing.T) {
	evidence := []string{"out/report.json"}
	uploads := []string{"out"}
	if countUploadedEvidenceOut(evidence, uploads) != 1 {
		t.Fatalf("directory upload did not bind evidence-out")
	}
}
