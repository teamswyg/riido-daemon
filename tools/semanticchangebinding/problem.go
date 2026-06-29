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
	lines := make([]string, 0, len(problems))
	for _, p := range problems {
		lines = append(lines, p.Message)
	}
	return errors.New(strings.Join(lines, "\n"))
}

func problemMessages(problems []problem) []string {
	out := make([]string, 0, len(problems))
	for _, p := range problems {
		out = append(out, p.Message)
	}
	return out
}
