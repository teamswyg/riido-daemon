package main

import (
	"reflect"
	"slices"
)

type claimDigest struct {
	ID        string        `json:"id"`
	Text      string        `json:"text"`
	Files     []string      `json:"files"`
	Docs      []string      `json:"docs"`
	Evidence  []string      `json:"evidence"`
	Verifiers []checkDigest `json:"verifiers"`
	Contracts []checkDigest `json:"contracts"`
}

type checkDigest struct {
	Name     string   `json:"name"`
	File     string   `json:"file"`
	Contains []string `json:"contains"`
}

func sameClaimSemantics(left, right businessClaim) bool {
	return reflect.DeepEqual(claimSemanticDigest(left), claimSemanticDigest(right))
}

func claimSemanticDigest(claim businessClaim) claimDigest {
	return claimDigest{
		ID:        claim.ID,
		Text:      claim.Text,
		Files:     sortedCopy(claim.Files),
		Docs:      sortedCopy(claim.Docs),
		Evidence:  sortedCopy(claim.Evidence),
		Verifiers: checkDigests(claim.Verifiers),
		Contracts: checkDigests(claim.Contracts),
	}
}

func sortedCopy(values []string) []string {
	out := append([]string{}, values...)
	slices.Sort(out)
	return out
}
