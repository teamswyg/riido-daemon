package main

import (
	"strings"
	"testing"
)

func TestValidateContractIdentityCoversValidAndMissingFields(t *testing.T) {
	valid := validContract()
	if problems := validateContractIdentity(valid); len(problems) != 0 {
		t.Fatalf("valid identity should pass: %v", problems)
	}

	invalid := valid
	invalid.SchemaVersion = "old"
	invalid.Product = " "
	invalid.ProviderCLIBundling = "allowed"

	problems := strings.Join(validateContractIdentity(invalid), "\n")
	for _, wanted := range []string{
		"schema_version must be",
		"product is required",
		`provider_cli_bundling must be "forbidden"`,
	} {
		if !strings.Contains(problems, wanted) {
			t.Fatalf("missing %q in %q", wanted, problems)
		}
	}
}

func TestValidateProviderCLINamesCoversEmptyValidAndInvalidNames(t *testing.T) {
	if got := validateProviderCLINames(nil); !hasError(got, "external_provider_cli_names must not be empty") {
		t.Fatalf("expected empty provider CLI problem, got %v", got)
	}
	if got := validateProviderCLINames([]string{"claude", "cursor-agent"}); len(got) != 0 {
		t.Fatalf("valid provider CLI names should pass: %v", got)
	}

	problems := strings.Join(validateProviderCLINames([]string{" ", "bin/claude", `bin\codex`}), "\n")
	for _, wanted := range []string{
		`invalid provider CLI name " "`,
		`invalid provider CLI name "bin/claude"`,
		`invalid provider CLI name "bin\\codex"`,
	} {
		if !strings.Contains(problems, wanted) {
			t.Fatalf("missing %q in %q", wanted, problems)
		}
	}
}

func TestChannelRequiredHelpersPreservePresentValuesAndReportMissing(t *testing.T) {
	var problems []string
	problems = appendRequiredString(problems, "value", "missing")
	problems = appendRequiredList(problems, "channel", "items", []string{"item"})
	if len(problems) != 0 {
		t.Fatalf("present fields should not append problems: %v", problems)
	}

	problems = appendRequiredField(problems, "channel", "field", " ")
	problems = appendRequiredList(problems, "channel", "items", nil)
	if !hasError(problems, `channel "channel" field is required`) {
		t.Fatalf("missing field problem not recorded: %v", problems)
	}
	if !hasError(problems, `channel "channel" items must not be empty`) {
		t.Fatalf("missing list problem not recorded: %v", problems)
	}
}
