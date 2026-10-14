package main

import "strings"

type problemError []string

func (err problemError) Error() string {
	return strings.Join(err, "\n")
}
