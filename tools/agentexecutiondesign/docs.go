package main

func renderedDocs(m model) map[string]string {
	return map[string]string{
		m.Manifest.GeneratedDoc:                    rootDoc(m),
		baseDir + "overview.md":                    overviewDoc(m),
		baseDir + "problem-map.md":                 problemMapDoc(m),
		baseDir + "current-structure-evidence.md":  currentStructureDoc(m),
		baseDir + "execution-identity.md":          executionIdentityDoc(m),
		baseDir + "workspace-plan.md":              workspacePlanDoc(m),
		baseDir + "runtime-launch-envelope.md":     launchEnvelopeDoc(m),
		baseDir + "stream-envelope.md":             streamEnvelopeDoc(m),
		baseDir + "retry-recovery-policy.md":       retryPolicyDoc(m),
		baseDir + "repo-ownership.md":              repoOwnershipDoc(m),
		baseDir + "implementation-slices.md":       implementationSlicesDoc(m),
		baseDir + "verification-evidence.md":       verificationEvidenceDoc(m),
		baseDir + "current-daemon-slice-status.md": currentStatusDoc(m),
		baseDir + "rag-guardrails.md":              ragGuardrailsDoc(m),
	}
}
