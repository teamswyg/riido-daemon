package main

type overviewFragment struct {
	SharedShape  []string  `json:"shared_shape"`
	FocusedFiles []linkRef `json:"focused_files"`
}

type riskFragment struct {
	Problems              []problemRow     `json:"problems"`
	StructureObservations []observationRow `json:"structure_observations"`
}

type executionFragment struct {
	IdentityFields  []fieldMeaning `json:"identity_fields"`
	IdentityRules   []string       `json:"identity_rules"`
	WorkspaceFields []phaseRule    `json:"workspace_fields"`
	LaunchFields    []ownerRule    `json:"launch_fields"`
}

type lifecycleFragment struct {
	StreamEvents         []streamEvent `json:"stream_events"`
	StreamRule           string        `json:"stream_rule"`
	RetryPolicies        []retryPolicy `json:"retry_policies"`
	ImplementationSlices []sliceSpec   `json:"implementation_slices"`
}

type governanceFragment struct {
	RepoOwnership []repoOwner `json:"repo_ownership"`
	RAGAllowed    []string    `json:"rag_allowed"`
	RAGForbidden  []string    `json:"rag_forbidden"`
	OpenDecisions []decision  `json:"open_decisions"`
	NonGoals      []string    `json:"non_goals"`
}
