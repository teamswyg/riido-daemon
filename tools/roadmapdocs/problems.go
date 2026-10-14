package main

import (
	"fmt"
	"strings"
)

func problemError(problems []string) error {
	return fmt.Errorf("%s", strings.Join(problems, "\n"))
}

func statusFor(problems []string) string {
	if len(problems) > 0 {
		return "failed"
	}
	return "verified"
}
