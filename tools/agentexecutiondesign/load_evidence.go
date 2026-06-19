package main

import "path/filepath"

func loadEvidence(repo string, out *model) error {
	if err := readJSON(repoPath(repo, out.Manifest.EvidenceManifest), &out.Evidence); err != nil {
		return err
	}
	base := filepath.Dir(out.Manifest.EvidenceManifest)
	for _, rel := range out.Evidence.EvidenceFiles.Local {
		if err := appendEvidenceFile(repo, base, rel, &out.Items); err != nil {
			return err
		}
	}
	for _, rel := range out.Evidence.EvidenceFiles.External {
		if err := appendEvidenceFile(repo, base, rel, &out.Items); err != nil {
			return err
		}
	}
	for _, rel := range out.Evidence.EvidenceFiles.RemainingBoundaries {
		if err := appendBoundaryFile(repo, base, rel, &out.Boundaries); err != nil {
			return err
		}
	}
	return nil
}

func appendEvidenceFile(repo, base, rel string, items *[]evidenceItem) error {
	var loaded []evidenceItem
	if err := readJSON(repoPath(repo, filepath.Join(base, rel)), &loaded); err != nil {
		return err
	}
	*items = append(*items, loaded...)
	return nil
}

func appendBoundaryFile(repo, base, rel string, items *[]boundaryItem) error {
	var loaded []boundaryItem
	if err := readJSON(repoPath(repo, filepath.Join(base, rel)), &loaded); err != nil {
		return err
	}
	*items = append(*items, loaded...)
	return nil
}
