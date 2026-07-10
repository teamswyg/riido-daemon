package main

import (
	"path/filepath"
	"testing"
)

func newFixture(t *testing.T) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.json")
	docPath := filepath.Join(dir, "doc.md")
	mustWrite(t, filepath.Join(dir, "workflow.yml"), "name: test\n")
	mustWrite(t, manifestPath, fixtureManifest())
	return dir, manifestPath, docPath
}

func fixtureManifest() string {
	return `{"schema_version":"riido-provider-real-cli-observation.v1","id":"test","title":"Test","generated_doc":"doc.md","workflow":"workflow.yml","evidence_artifact":"artifact","execution_policy":{"public_runner_mode":"contract_only","paid_runner_mode":"local_private_mac","cadence":"weekly","timezone":"UTC","runs_per_month":4},"budget_policy":{"monthly_cash_budget_usd_per_provider":20,"max_estimated_input_tokens_per_run":1000,"max_estimated_output_tokens_per_run":500,"max_estimated_total_tokens_per_provider_month":6000,"fail_closed":true,"usage_accounting":"estimated"},"scenarios":[{"id":"side-effect","execution":"integration_test","paid":true,"estimated_input_tokens":1000,"estimated_output_tokens":500,"max_duration_seconds":60}],"failure_promotion":{"candidate_producer":".","decision_verifier":".","automation_mode":"test","failure_states":["failed"],"merge_policy":"pull_request_ci"},"providers":[{"id":"fake","display_name":"Fake","default_executable":"missing-riido-provider","override_env":"RIIDO_FAKE_PROVIDER_PATH","go_package":".","test_regex":"TestIntegration","scenario_ids":["side-effect"]}]}`
}
