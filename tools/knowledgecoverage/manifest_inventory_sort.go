package main

import "sort"

func manifestGroups(counts map[string]int) []manifestGroupCount {
	groups := make([]manifestGroupCount, 0, len(counts))
	for group, count := range counts {
		groups = append(groups, manifestGroupCount{Group: group, Count: count})
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Count == groups[j].Count {
			return groups[i].Group < groups[j].Group
		}
		return groups[i].Count > groups[j].Count
	})
	return groups
}

func manifestSamples(groups []manifestGroupCount, samples map[string][]string) []manifestGroupSample {
	out := make([]manifestGroupSample, 0, len(groups))
	for _, group := range groups {
		out = append(out, manifestGroupSample{Group: group.Group, Paths: samples[group.Group]})
	}
	return out
}
