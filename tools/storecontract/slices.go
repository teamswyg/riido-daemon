package main

import "slices"

func contains(items []string, wanted string) bool {
	return slices.Contains(items, wanted)
}
