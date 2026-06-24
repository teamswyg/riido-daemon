package main

func (p changedProblem) summary() string {
	return p.ClaimID + " changed runtime files without bound doc/verifier/registry evidence"
}

func firstChangedFile(p changedProblem) string {
	if len(p.ChangedFiles) == 0 {
		return defaultManifest
	}
	return p.ChangedFiles[0]
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
