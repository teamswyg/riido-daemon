package main

import "slices"

func intersects(left, right []string) bool {
	for _, item := range left {
		if slices.Contains(right, item) {
			return true
		}
	}
	return false
}

func intersection(left, right []string) []string {
	var out []string
	for _, item := range left {
		if slices.Contains(right, item) {
			out = append(out, item)
		}
	}
	return out
}
