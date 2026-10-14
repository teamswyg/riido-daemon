package main

import "testing"

func TestArtifactUploadModes(t *testing.T) {
	text := "" +
		"steps:\n" +
		"  - uses: actions/upload-artifact@v4\n" +
		"    with:\n" +
		"      if-no-files-found: error\n" +
		"  - uses: actions/upload-artifact@v4\n" +
		"    with:\n" +
		"      name: optional\n"
	modes := artifactUploadModes(text)
	if len(modes) != 2 || modes[0] != "error" || modes[1] != "" {
		t.Fatalf("unexpected modes: %#v", modes)
	}
}

func TestWorkflowEvidenceRejectsWarnUpload(t *testing.T) {
	record := workflowRecord{
		Path:                 ".github/workflows/example.yml",
		HasExecutable:        true,
		HasEvidenceOut:       true,
		UploadsArtifact:      true,
		ArtifactUploadCount:  1,
		NonStrictUploadCount: 1,
	}
	got := classify(record, nil, nil)
	if got.Status != "non_strict_upload" {
		t.Fatalf("expected non_strict_upload, got %q", got.Status)
	}
}
