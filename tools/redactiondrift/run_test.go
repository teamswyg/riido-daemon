package main

import "testing"

func TestCurrentRepoHasNoRedactionDrift(t *testing.T) {
	if err := run("../.."); err != nil {
		t.Fatal(err)
	}
}

func TestSecurityHubMustLinkRedactionSSOT(t *testing.T) {
	problems := validateSecurityHub("docs/20-domain/security/example.md", "log redaction is required")
	if len(problems) != 1 {
		t.Fatalf("problems = %v, want one missing-link problem", problems)
	}
}

func TestSecurityHubCannotRedefineMarkerLiteral(t *testing.T) {
	text := "[`../security-redaction.md`](../security-redaction.md) says use [redacted]"
	problems := validateSecurityHub("docs/20-domain/security/example.md", text)
	if len(problems) != 1 {
		t.Fatalf("problems = %v, want one marker redefinition problem", problems)
	}
}
