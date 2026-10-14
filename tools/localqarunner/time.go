package main

import (
	"strings"
	"time"
)

func timestampSlug(observedAt string) string {
	t, err := time.Parse(time.RFC3339, observedAt)
	if err != nil {
		return strings.NewReplacer("-", "", ":", "").Replace(observedAt)
	}
	return t.UTC().Format("20060102T150405Z")
}
