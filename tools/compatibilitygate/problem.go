package main

import (
	"fmt"
	"strings"
)

type problem struct {
	Message string
}

func problemError(problems []problem) error {
	var b strings.Builder
	for _, p := range problems {
		fmt.Fprintf(&b, "%s\n", p.Message)
	}
	return fmt.Errorf("%s", strings.TrimSpace(b.String()))
}

func problemMessages(problems []problem) []string {
	out := make([]string, 0, len(problems))
	for _, p := range problems {
		out = append(out, p.Message)
	}
	return out
}
