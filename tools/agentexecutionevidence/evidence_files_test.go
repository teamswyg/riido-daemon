package agentexecutionevidence

import (
	"path/filepath"
	"testing"
)

func collectLocalEvidence(t *testing.T, manifest evidenceManifest, path string) []localEvidence {
	t.Helper()
	out := append([]localEvidence{}, manifest.LocalEvidence...)
	for _, file := range manifest.EvidenceFiles.Local {
		out = append(out, loadJSONFile[[]localEvidence](t, evidenceFilePath(path, file))...)
	}
	return out
}

func collectExternalEvidence(t *testing.T, manifest evidenceManifest, path string) []externalEvidence {
	t.Helper()
	out := append([]externalEvidence{}, manifest.ExternalEvidence...)
	for _, file := range manifest.EvidenceFiles.External {
		out = append(out, loadJSONFile[[]externalEvidence](t, evidenceFilePath(path, file))...)
	}
	return out
}

func collectRemainingBoundaries(t *testing.T, manifest evidenceManifest, path string) []remainingBoundary {
	t.Helper()
	out := append([]remainingBoundary{}, manifest.RemainingBoundaries...)
	for _, file := range manifest.EvidenceFiles.RemainingBoundaries {
		out = append(out, loadJSONFile[[]remainingBoundary](t, evidenceFilePath(path, file))...)
	}
	return out
}

func evidenceFilePath(manifestPath, file string) string {
	return filepath.Join(filepath.Dir(manifestPath), filepath.FromSlash(file))
}
