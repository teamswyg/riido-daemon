package main

import (
	"fmt"
	"strings"
)

type problem struct {
	Message string `json:"message"`
}

type problemList []problem

func (p problemList) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d full-access harness evidence problem(s):", len(p))
	for _, item := range p {
		fmt.Fprintf(&b, "\n- %s", item.Message)
	}
	return b.String()
}

func problemError(problems []problem) error {
	return problemList(problems)
}
