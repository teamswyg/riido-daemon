package main

import "strings"

func mentionsRedaction(text string) bool {
	lower := strings.ToLower(text)
	for _, term := range redactionTerms {
		if strings.Contains(lower, strings.ToLower(term)) {
			return true
		}
	}
	return false
}

func hasSSOTLink(text string) bool {
	for _, link := range ssotLinks {
		if strings.Contains(text, link) {
			return true
		}
	}
	return false
}
