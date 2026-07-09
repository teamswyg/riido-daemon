package main

import (
	"strings"
	"testing"
)

func TestValidateManifestReportsSecretContractDrift(t *testing.T) {
	dir := t.TempDir()
	m := manifest{
		SchemaVersion:        "bad",
		Workflow:             "missing.yml",
		AllowedPacketFields:  []string{"token"},
		ForbiddenFieldNames:  []string{"token"},
		AllowedAWSOperations: []string{"ssm:GetParameter"},
		ForbiddenAWSOps:      []string{"ssm:GetParameter"},
		EvidenceKinds: []evidenceKind{
			{ID: "same"},
			{ID: "same", ActualID: "actual", Proves: []string{"metadata"}, Forbids: []string{"secret"}},
		},
	}
	problems := validateManifest(dir, m)
	for _, want := range []string{
		"schema_version must be riido-runtime-secret-private-evidence.v1",
		"id, title, generated_doc, workflow, evidence_artifact, and private_owner are required",
		"public_scope must not be empty",
		"missing workflow \"missing.yml\"",
		"same must include proves and forbids",
		"duplicate evidence kind \"same\"",
		"forbidden field \"token\" is also allowed",
		"forbidden_field_names missing \"secret\"",
		"forbidden AWS operation \"ssm:GetParameter\" is also allowed",
		"forbidden_aws_operations missing \"kms:Decrypt\"",
		"allowed_aws_operations must include ssm:DescribeParameters",
	} {
		assertSecretProblem(t, problems, want)
	}
}

func TestBuildEvidenceCopiesOnlyMetadataShape(t *testing.T) {
	m := manifest{
		ID:                   "runtime-secret",
		PrivateOwner:         "private-infra",
		AllowedAWSOperations: []string{"ssm:DescribeParameters"},
		ForbiddenAWSOps:      []string{"ssm:GetParameter"},
		AllowedPacketFields:  []string{"name"},
		ForbiddenFieldNames:  []string{"token"},
	}
	ev := buildEvidence(m)
	if ev.ID != "runtime-secret" || ev.Status != "verified" || ev.PrivateOwner != "private-infra" {
		t.Fatalf("unexpected evidence header %#v", ev)
	}
	if len(ev.Assertions) == 0 || ev.ForbiddenFields[0] != "token" ||
		ev.AllowedOperations[0] != "ssm:DescribeParameters" {
		t.Fatalf("metadata evidence not copied %#v", ev)
	}
}

func assertSecretProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, problem := range problems {
		if strings.Contains(problem, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}
