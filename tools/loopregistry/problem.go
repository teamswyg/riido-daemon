package main

import "strings"

func joinProblems(items []string) string {
	return strings.Join(items, "\n- ")
}
