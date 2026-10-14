package main

import (
	"errors"
	"strings"
)

type problem struct {
	Message string `json:"message"`
}

func problemError(problems []problem) error {
	if len(problems) == 0 {
		return nil
	}
	lines := make([]string, 0, len(problems))
	for _, item := range problems {
		lines = append(lines, item.Message)
	}
	return errors.New(strings.Join(lines, "\n"))
}
