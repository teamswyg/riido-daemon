package main

var redactionTerms = []string{
	"redaction",
	"redacted",
	"redact",
	"secret 패턴",
	"금지 패턴",
}

var ssotLinks = []string{
	"security-redaction.md",
	"../security-redaction.md",
	"../../security-redaction.md",
}

var forbiddenOutsideSSOT = []string{
	"[REDACTED:",
	"[redacted]",
	"Canonical IR payload redaction marker",
	"ToolRef.Args marker",
}
