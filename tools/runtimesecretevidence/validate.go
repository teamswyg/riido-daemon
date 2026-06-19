package main

import (
	"fmt"
	"os"
)

var requiredForbiddenOps = []string{
	"ssm:GetParameter",
	"ssm:GetParameters",
	"ssm:GetParametersByPath",
	"kms:Decrypt",
	"secretsmanager:GetSecretValue",
}

var requiredForbiddenFields = []string{
	"value", "token", "secret", "secret_string", "payload",
	"parameter_value", "decrypted_value", "authorization", "bearer",
}

func validateManifest(root string, m manifest) []string {
	var problems []string
	problems = append(problems, validateHeader(root, m)...)
	problems = append(problems, validateKinds(m.EvidenceKinds)...)
	problems = append(problems, validatePacketFields(m)...)
	problems = append(problems, validateAWSOperations(m)...)
	return problems
}

func validateHeader(root string, m manifest) []string {
	var problems []string
	if m.SchemaVersion != "riido-runtime-secret-private-evidence.v1" {
		problems = append(problems, "schema_version must be riido-runtime-secret-private-evidence.v1")
	}
	for _, value := range []string{m.ID, m.Title, m.GeneratedDoc, m.Workflow, m.EvidenceArtifact, m.PrivateOwner} {
		if value == "" {
			problems = append(problems, "id, title, generated_doc, workflow, evidence_artifact, and private_owner are required")
		}
	}
	if len(m.PublicScope) == 0 {
		problems = append(problems, "public_scope must not be empty")
	}
	if _, err := os.Stat(resolvePath(root, m.Workflow)); err != nil {
		problems = append(problems, fmt.Sprintf("missing workflow %q", m.Workflow))
	}
	return problems
}
