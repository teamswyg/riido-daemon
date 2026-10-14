package main

func defaultCommand() string {
	return "go run ./tools/loopregistry -check-doc -evidence-out out/loop-registry.json"
}

func fixtureManifest() string {
	return `{
  "schema_version": "riido-loop-registry.v1",
  "id": "fixture",
  "title": "Fixture Loop Registry",
  "generated_doc": "docs/30-architecture/loop-registry.md",
  "workflow": ".github/workflows/loop-registry.yml",
  "evidence_artifact": "loop-registry",
  "precommit_hook": "loop-registry",
  "command": "` + defaultCommand() + `",
  "loop": {
    "observation": "o",
    "hypothesis": "h",
    "execute": "x",
    "evaluate": "e",
    "retrospective": "r"
  },
  "loops": [{
    "id": "claim-loop",
    "owner": "fixture",
    "kind": "closed-loop",
    "observes": ["code.go"],
    "verifies": ["TestClaimBinding"],
    "evidence": ["out/loop-registry.json"],
    "expires_after": "24h",
    "fails_when": ["semantic_drift"],
    "evidence_graph": {
      "observation": "o",
      "hypothesis": "h",
      "change": "c",
      "verifier": "v",
      "evidence": "e",
      "decision": "d",
      "next_loop": "n"
    }
  }],
  "business_claims": [{
    "id": "claim_binding",
    "text": "Claims bind code, docs, and verifier tokens.",
    "files": ["code.go"],
    "docs": ["doc.md"],
    "evidence": ["out/loop-registry.json"],
    "verifiers": [{"name":"claim-test","file":"code_test.go","contains":["TestClaimBinding"]}]
  }]
}`
}
