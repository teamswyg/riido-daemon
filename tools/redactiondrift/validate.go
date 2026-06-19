package main

import (
	"fmt"
	"strings"
)

func validateDoc(path, text string) []string {
	if isSecurityRedactionDoc(path) {
		return validateRedactionSSOT(path, text)
	}
	return validateSecurityHub(path, text)
}

func validateSecurityHub(path, text string) []string {
	var problems []string
	if mentionsRedaction(text) && !hasSSOTLink(text) {
		problems = append(problems, fmt.Sprintf("%s mentions redaction without linking security-redaction.md", path))
	}
	for _, literal := range forbiddenOutsideSSOT {
		if strings.Contains(text, literal) {
			problems = append(problems, fmt.Sprintf("%s redefines redaction literal %q", path, literal))
		}
	}
	return problems
}

func validateRedactionSSOT(path, text string) []string {
	if path != "docs/20-domain/security-redaction/markers.md" {
		return nil
	}
	var problems []string
	for _, literal := range []string{"[REDACTED:<patternID>]", "[redacted]"} {
		if !strings.Contains(text, literal) {
			problems = append(problems, fmt.Sprintf("%s missing marker literal %q", path, literal))
		}
	}
	return problems
}
