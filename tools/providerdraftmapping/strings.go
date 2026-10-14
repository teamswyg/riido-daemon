package main

import "strings"

func camel(value string) string {
	var b strings.Builder
	for part := range strings.SplitSeq(value, "_") {
		if part == "" {
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]))
		b.WriteString(part[1:])
	}
	return b.String()
}
