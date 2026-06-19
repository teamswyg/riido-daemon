package main

func worktreeStatus(p providerEvidence) string {
	if p.Provider == "openclaw" {
		return "`supports_worktree=false`; `required_surfaces=[worktree]` must fail with `MISSING_REQUIRED_SURFACE:worktree`"
	}
	if p.WorktreeSupport == "supported" {
		return "`supports_worktree=true`"
	}
	return "`supports_worktree=false`"
}

func realProviderByID(m manifest, id string) (realProvider, bool) {
	for _, provider := range m.RealCLIObservation.Providers {
		if provider.ID == id {
			return provider, true
		}
	}
	return realProvider{}, false
}
