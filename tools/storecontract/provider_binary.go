package main

import (
	"path/filepath"
	"strings"
)

func matchesProviderBinary(filename string, providerNames []string) bool {
	base := strings.ToLower(filename)
	ext := strings.ToLower(filepath.Ext(base))
	if !isExecutableExtension(ext) {
		return false
	}
	stem := strings.TrimSuffix(base, ext)
	for _, provider := range providerNames {
		name := strings.ToLower(provider)
		if base == name || stem == name {
			return true
		}
	}
	return false
}

func isExecutableExtension(ext string) bool {
	switch ext {
	case "", ".exe", ".cmd", ".bat", ".ps1", ".sh":
		return true
	default:
		return false
	}
}
