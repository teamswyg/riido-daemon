package main

func testModel() model {
	return model{
		Manifest: manifest{
			SchemaVersion:    "riido-agent-execution-design-docs.v1",
			ID:               "agent-execution-design-docs-test",
			Title:            "Agent Execution Design Test",
			GeneratedDoc:     "docs/root.md",
			Workflow:         ".github/workflows/agent-execution-design-docs.yml",
			EvidenceArtifact: "out/agent-execution-design-docs.json",
			RiidoTask:        "RIID-test",
			AssignmentFSMDoc: "assignment-lifecycle-fsm.md",
			EvidenceManifest: "docs/evidence.riido.json",
		},
		Overview: overviewFragment{
			SharedShape:  []string{"observe", "verify"},
			FocusedFiles: []linkRef{{Title: "Overview", Path: "overview.md"}},
		},
		Risk: riskFragment{
			Problems: []problemRow{{
				ID: "risk-one", Symptom: "symptom", Cause: "cause", Direction: "direction",
			}},
			StructureObservations: []observationRow{{
				Observation: "obs", SSOT: "ssot.json", Meaning: "meaning",
			}},
		},
		Execution: executionFragment{
			IdentityFields:  []fieldMeaning{{Field: "assignment_id", Meaning: "execution key"}},
			IdentityRules:   []string{"assignment id is stable"},
			WorkspaceFields: []phaseRule{{Field: "repo", Phase: "prepare", Rule: "clone"}},
			LaunchFields:    []ownerRule{{Field: "argv", Owner: "daemon", Rule: "locked"}},
		},
		Lifecycle: lifecycleFragment{
			StreamEvents: []streamEvent{{
				Kind: "progress", Store: "journal", ClientMeaning: "line",
			}},
			StreamRule:           "terminal does not revive",
			RetryPolicies:        []retryPolicy{{Class: "auth", Retry: "no", Rule: "fail closed"}},
			ImplementationSlices: []sliceSpec{{Title: "slice", Items: []string{"item"}}},
		},
		Governance: governanceFragment{
			RepoOwnership: []repoOwner{{Repo: "riido-daemon", Responsibility: "runtime"}},
			RAGAllowed:    []string{"public docs"},
			RAGForbidden:  []string{"secrets"},
			OpenDecisions: []decision{{ID: "d1", Decision: "choose", Default: "deny"}},
			NonGoals:      []string{"private clone"},
		},
		Items: []evidenceItem{{
			Risk: "risk-one", Status: "verified", Repo: "riido-daemon", Test: "TestOne", Proves: "proof",
		}},
		Boundaries: []boundaryItem{{
			ID: "b1", Owner: "daemon", CurrentHandling: "manual", RequiredNextArtifact: "verifier",
		}},
	}
}
