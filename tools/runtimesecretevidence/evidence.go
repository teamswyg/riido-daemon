package main

import "time"

func buildEvidence(m manifest) evidenceFile {
	return evidenceFile{
		SchemaVersion:       "riido-runtime-secret-private-evidence-result.v1",
		ID:                  m.ID,
		ObservedAt:          time.Now().UTC().Format(time.RFC3339),
		Status:              "verified",
		PrivateOwner:        m.PrivateOwner,
		AllowedOperations:   m.AllowedAWSOperations,
		ForbiddenOperations: m.ForbiddenAWSOps,
		AllowedFields:       m.AllowedPacketFields,
		ForbiddenFields:     m.ForbiddenFieldNames,
		Assertions: []string{
			"public daemon repo owns metadata-only contract shape",
			"raw runtime secret values are not valid packet fields",
			"secret read/decrypt AWS operations are forbidden by public contract",
			"private infra evidence must be attached as sanitized metadata only",
		},
	}
}
