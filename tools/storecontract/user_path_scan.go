package main

import (
	"os"
	"strings"
)

var hardcodedUserPathMarkers = []string{
	"/Users/",
	`C:\Users\`,
	"C:/Users/",
	"~/Library/LaunchAgents",
	"~/Library/Application Support",
}

func hasHardcodedUserPath(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	text := string(data)
	for _, marker := range hardcodedUserPathMarkers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}
