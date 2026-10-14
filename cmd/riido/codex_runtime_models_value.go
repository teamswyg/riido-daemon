package main

import (
	"strconv"
	"strings"
)

func parseCodexModelValue(rawValue string) string {
	value := strings.TrimSpace(rawValue)
	if unquoted, err := strconv.Unquote(value); err == nil {
		return strings.TrimSpace(unquoted)
	}
	if commentAt := strings.Index(value, "#"); commentAt >= 0 {
		value = strings.TrimSpace(value[:commentAt])
		if unquoted, err := strconv.Unquote(value); err == nil {
			return strings.TrimSpace(unquoted)
		}
	}
	return strings.TrimSpace(value)
}
