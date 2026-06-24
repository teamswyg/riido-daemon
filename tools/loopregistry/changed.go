package main

import (
	"os"
	"slices"
	"strings"
)

func changedCheck(root string, reg registry, path string) changedSummary {
	changed, err := readChangedFiles(root, path)
	if err != nil {
		return changedSummary{Problems: []string{err.Error()}}
	}
	out := changedSummary{InputCount: len(changed)}
	for _, claim := range reg.BusinessClaims {
		if !intersects(changed, claim.Files) {
			continue
		}
		out.MatchedClaims = append(out.MatchedClaims, claim.ID)
		out.Problems = append(out.Problems, validateClaimChange(claim, changed)...)
	}
	out.MatchedClaimCount = len(out.MatchedClaims)
	return out
}

func readChangedFiles(root, rel string) ([]string, error) {
	body, err := os.ReadFile(repoPath(root, rel))
	if err != nil {
		return nil, err
	}
	var out []string
	for line := range strings.SplitSeq(string(body), "\n") {
		if trimmed := slash(line); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out, nil
}

func validateClaimChange(claim businessClaim, changed []string) []string {
	if !intersects(changed, claim.Files) {
		return nil
	}
	if intersects(changed, append(claimDocs(claim), registryFiles()...)) {
		return nil
	}
	return []string{claim.ID + " changed runtime files without bound doc/verifier/registry evidence"}
}

func claimDocs(claim businessClaim) []string {
	out := append([]string{}, claim.Docs...)
	for _, check := range append(claim.Verifiers, claim.Contracts...) {
		out = append(out, check.File)
	}
	return out
}

func registryFiles() []string {
	return []string{defaultManifest, "docs/30-architecture/loop-registry.md"}
}

func intersects(left, right []string) bool {
	for _, item := range left {
		if slices.Contains(right, item) {
			return true
		}
	}
	return false
}
