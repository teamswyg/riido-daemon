package main

import "slices"

func checkDigests(checks []sourceCheck) []checkDigest {
	out := make([]checkDigest, 0, len(checks))
	for _, check := range checks {
		out = append(out, checkDigest{
			Name:     check.Name,
			File:     check.File,
			Contains: sortedCopy(check.Contains),
		})
	}
	slices.SortFunc(out, compareCheckDigest)
	return out
}

func compareCheckDigest(left, right checkDigest) int {
	switch {
	case left.Name != right.Name:
		return compareString(left.Name, right.Name)
	case left.File != right.File:
		return compareString(left.File, right.File)
	default:
		return slices.Compare(left.Contains, right.Contains)
	}
}

func compareString(left, right string) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
