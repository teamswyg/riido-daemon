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
	b.WriteString("tool use gate evidence failed:")
	for _, p := range problems {
		b.WriteString("\n- ")
		b.WriteString(p.Message)
	}
	return errors.New(b.String())
}
