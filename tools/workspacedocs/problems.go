package main

import "strings"

func problemText(problems []string) string {
	if len(problems) == 0 {
		return ""
	}
	return strings.Join(problems, "\n")
}
