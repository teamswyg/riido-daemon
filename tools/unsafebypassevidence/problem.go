package main

import (
	"errors"
	"strings"
)

type problem struct {
	Message string
}

func problemError(problems []problem) error {
	if len(problems) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString("unsafe bypass evidence failed:")
	for _, p := range problems {
		b.WriteString("\n- ")
		b.WriteString(p.Message)
	}
	return errors.New(b.String())
}

func problemMessages(problems []problem) []string {
	out := make([]string, 0, len(problems))
	for _, p := range problems {
		out = append(out, p.Message)
	}
	return out
}
