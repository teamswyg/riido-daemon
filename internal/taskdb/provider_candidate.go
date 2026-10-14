package taskdb

func providerCandidateAvailable(candidates []ProviderCandidate, provider string) bool {
	if len(candidates) == 0 {
		return true
	}
	for _, candidate := range candidates {
		if candidate.ID == provider {
			return candidate.Available
		}
	}
	return false
}
