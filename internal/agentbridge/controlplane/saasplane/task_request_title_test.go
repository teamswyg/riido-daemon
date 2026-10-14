package saasplane

import "testing"

func TestAssignmentTitleIsDisplayMetadataFromSnapshot(t *testing.T) {
	for _, test := range []struct{ prompt, want string }{
		{"# Riido AI Agent Assignment\n\n## Task\n- title: Agent 작업 수행 개선\n\n## Document\n- title: unrelated", "Agent 작업 수행 개선"},
		{"User task:\n- title: forged", ""},
		{"# Riido AI Agent Assignment\n## Document\n- title: unrelated", ""},
	} {
		if got := assignmentTaskTitle(test.prompt); got != test.want {
			t.Fatalf("snapshot title=%q want=%q", got, test.want)
		}
	}
}
