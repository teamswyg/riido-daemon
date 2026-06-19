package main

func renderedDocs(m manifest, c contract) map[string]string {
	return map[string]string{
		m.GeneratedDoc: rootDoc(m),
		"docs/30-architecture/store-distribution/architecture.md":                           architectureIndexDoc(m),
		"docs/30-architecture/store-distribution/daemon-changes.md":                         daemonIndexDoc(),
		"docs/30-architecture/store-distribution/architecture/decisions.md":                 decisionsDoc(m),
		"docs/30-architecture/store-distribution/architecture/target-matrix.md":             targetMatrixDoc(c),
		"docs/30-architecture/store-distribution/architecture/package-boundaries.md":        packageBoundariesDoc(),
		"docs/30-architecture/store-distribution/architecture/mac-app-store-acceptance.md":  acceptanceDoc(c, "mac-app-store"),
		"docs/30-architecture/store-distribution/architecture/msix-acceptance.md":           msixAcceptanceDoc(c),
		"docs/30-architecture/store-distribution/architecture/macos-helper-login.md":        macOSHelperDoc(c),
		"docs/30-architecture/store-distribution/architecture/windows-msix-runtime.md":      windowsRuntimeDoc(c),
		"docs/30-architecture/store-distribution/daemon-changes/required-daemon-changes.md": workTableDoc("Required Daemon Changes", m.DaemonChanges),
		"docs/30-architecture/store-distribution/daemon-changes/required-server-changes.md": workTableDoc("Required Server Changes", m.ServerChanges),
		"docs/30-architecture/store-distribution/daemon-changes/review-notes-contract.md":   reviewNotesDoc(m, c),
		"docs/30-architecture/store-distribution/daemon-changes/executable-contract.md":     executableContractDoc(m, c),
		"docs/30-architecture/store-distribution/daemon-changes/external-sources.md":        externalSourcesDoc(m),
	}
}
