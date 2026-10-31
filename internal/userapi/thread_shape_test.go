package userapi

import (
	"errors"
	"strings"
	"testing"
)

func TestThreadsRejectCaseAliasesAndNullRequiredFields(t *testing.T) {
	for name, raw := range map[string]string{
		"task alias":     strings.Replace(threadHistoryFixture, `"task_id":"task-one"`, `"task_id":"wrong","TASK_ID":"task-one"`, 1),
		"message alias":  strings.Replace(threadHistoryFixture, `"role":"agent"`, `"role":"invalid","ROLE":"agent"`, 1),
		"snapshot alias": strings.Replace(threadHistoryFixture, `"name":"에이전트"`, `"name":"original","NAME":"replacement"`, 1),
		"null run":       strings.Replace(threadHistoryFixture, `"run_id":"run-one"`, `"run_id":null`, 1),
		"null status":    strings.Replace(threadHistoryFixture, `"work_status":"done"`, `"work_status":null`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			r := ThreadReader{Client: &threadFixtureClient{body: raw, status: 200}}
			got, err := r.Show(t.Context(), "workspace-one", "task-one", "thread-one")
			if got != nil || !errors.Is(err, ErrInvalidResponse) {
				t.Fatal("ambiguous response accepted", err)
			}
		})
	}
}

func TestThreadsAllowDistinctCaseSensitiveSnapshotIDs(t *testing.T) {
	raw := strings.Replace(threadHistoryFixture, `"agent_snapshots":{`, `"agent_snapshots":{"Snapshot-one":{"agent_id":"agent-two","name":"Second"},`, 1)
	got, err := (ThreadReader{Client: &threadFixtureClient{body: raw, status: 200}}).List(t.Context(), "workspace-one", "task-one")
	if err != nil || len(got.AgentSnapshots) != 2 {
		t.Fatal("case-sensitive data keys rejected", err)
	}
}
