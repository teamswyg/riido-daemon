package main

import "sort"

type statusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

func workflowStatusCounts(records []workflowRecord) []statusCount {
	counts := map[string]int{}
	for _, record := range records {
		counts[record.Status]++
	}
	out := make([]statusCount, 0, len(counts))
	for status, count := range counts {
		out = append(out, statusCount{Status: status, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Status < out[j].Status
	})
	return out
}
