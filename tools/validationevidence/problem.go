package main

import (
	"fmt"
	"strings"
)

func problemError(problems []problem) error {
	var b strings.Builder
	b.WriteString("validation evidence invalid:\n")
	for _, item := range problems {
		fmt.Fprintf(&b, "- %s\n", item.Message)
	}
	return fmt.Errorf("%s", b.String())
}

func messages(problems []problem) []string {
	out := make([]string, 0, len(problems))
	for _, item := range problems {
		out = append(out, item.Message)
	}
	return out
}
