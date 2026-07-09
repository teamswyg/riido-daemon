package main

import (
	"strings"
	"testing"
)

func TestValidateCoreReportsRequiredAndInvariantIssues(t *testing.T) {
	core := coreDoc{Invariants: []invariant{{Name: "pin"}}}
	problems := validateCore(core)
	assertRuntimeSchedulingProblem(t, problems, "core id, title, generated_doc, and context are required")
	assertRuntimeSchedulingProblem(t, problems, "core responsibilities and non-responsibilities are required")
	assertRuntimeSchedulingProblem(t, problems, "nine core invariants are required")
	assertRuntimeSchedulingProblem(t, problems, "invariant name, summary, and source checks are required")
}

func TestValidateLinksAndInvariantRefsReportProblems(t *testing.T) {
	linkProblems := validateLinks("parts", []link{{Title: "T"}}, 2)
	assertRuntimeSchedulingProblem(t, linkProblems, "parts count = 1, want 2")
	linkProblems = validateLinks("parts", []link{{Title: "T"}}, 1)
	assertRuntimeSchedulingProblem(t, linkProblems, "parts links require title and path")
	m := manifest{
		SourceChecks: []sourceCheck{{Name: "known"}},
		Core:         coreDoc{Invariants: []invariant{{SourceChecks: []string{"missing"}}}},
	}
	assertRuntimeSchedulingProblem(t, validateInvariantChecks(m), "unknown invariant source check missing")
}

func TestValidateIndexReportsMissingFields(t *testing.T) {
	problems := validateIndex(indexDoc{Parts: []link{{Title: "A", Path: "a.md"}}})
	assertRuntimeSchedulingProblem(t, problems, "invariants index title, generated_doc, and summary are required")
	assertRuntimeSchedulingProblem(t, problems, "invariants parts count = 1, want 4")
}

func assertRuntimeSchedulingProblem(t *testing.T, problems []string, needle string) {
	t.Helper()
	for _, p := range problems {
		if strings.Contains(p, needle) {
			return
		}
	}
	t.Fatalf("missing problem %q in %#v", needle, problems)
}
