package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteContractLabRendersVisualEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index.html")
	manualOut := filepath.Join(t.TempDir(), "manual-qa-evidence.json")
	screenshots := filepath.Join(t.TempDir(), "screenshots")
	evidence := evidenceFile{
		Status:    statusPassed,
		ExpiresAt: "2999-06-23T00:00:00Z",
		Scenarios: []scenario{
			{
				ID:         "figma.onboarding",
				Status:     statusPassed,
				Screenshot: ".riido-local/screenshots/figma-onboarding.png",
			},
			{
				ID:     "figma.intent.catalog",
				Status: statusPassed,
				Observed: map[string]any{
					"entries_count": 1,
					"entries": []map[string]any{{
						"node_id":               "1179:27360",
						"name":                  "런타임 상세페이지",
						"daemon_scope":          "Projects one selected runtime.",
						"daemon_consumed_facts": []string{"runtime id"},
						"client_owned_facts":    []string{"agent row hover"},
					}},
				},
			},
			featureUIScenario(),
			qaI18NScenario(),
			browserMeaningScenario(),
			{
				ID:     "domain.fixture_journey",
				Status: statusPassed,
				Observed: map[string]any{
					"remote_environment":  "staging",
					"verification_source": "local",
					"cache_path":          ".riido-local/evidence/domain-fixture-journey-cache.json",
				},
			},
			{
				ID:     "domain.fixture.thread",
				Status: statusSkipped,
				Observed: map[string]any{
					"title":           "Thread",
					"create_endpoint": "POST /tasks/{task_id}/agent-assignments",
					"verify_endpoint": "GET /tasks/{task_id}/thread-stream-subscription",
				},
			},
			evidenceGapScenario(nil, config{manualOut: &manualOut, screenshots: &screenshots}),
		},
	}
	if err := writeContractLab(path, evidence); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	html := string(data)
	for _, want := range []string{
		"id=\"qa-i18n\"",
		"riido-qa-ui-i18n-dsl.v1",
		"시각 증적",
		"className: \"shot\"",
		"../screenshots/",
		"기능 영역",
		"증적 재생",
		"도메인 여정",
		"내부 실행 확인 화면",
		"페이지 {page} / {total}",
		"다음 단계",
		"Figma 증적",
		"수동 QA",
		"Figma 의도",
		"매일 QA 루프",
		"manual-qa-evidence.json",
		"riido-manual-qa-evidence.v1",
		"riido-domain-fixture-cache.v1",
		"domain-fixture-journey-cache.json",
		"contract.ui.feature_dsl",
		"contract.ui.i18n_dsl",
		"contract.ui.browser_meaning_qa",
		"domain.fixture_journey",
		"local.qa.evidence_gap_candidates",
		"manual-evidence-file",
		"max_pixels",
		"1179:27360",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("contract lab missing %q", want)
		}
	}
}
