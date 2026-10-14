package main

import "testing"

func TestSummarizeAssignedProfilesCountsTaskKeys(t *testing.T) {
	got := summarizeAssignedProfiles(map[string]any{
		"workspace_id": "workspace-a",
		"assigned_agent_profiles": map[string]any{
			"task-a": map[string]any{},
			"task-b": map[string]any{},
		},
	})
	if got["workspace_id_present"] != true {
		t.Fatalf("workspace_id_present = %v", got["workspace_id_present"])
	}
	if got["assigned_task_keys_count"] != 2 {
		t.Fatalf("assigned_task_keys_count = %v", got["assigned_task_keys_count"])
	}
}

func TestFirstAssignedProfileTaskIDIsDeterministic(t *testing.T) {
	got := firstAssignedProfileTaskID(map[string]any{
		"assigned_agent_profiles": map[string]any{
			"task-b": map[string]any{},
			"task-a": map[string]any{},
		},
	})
	if got != "task-a" {
		t.Fatalf("task=%q", got)
	}
}
