package main

import "path/filepath"

func buildReport(dir string, m manifest) report {
	out := newReport(m)
	for _, item := range m.Required {
		collectEvidence(&out, dir, item)
	}
	for _, check := range out.Checks {
		if check.Status == statusVerified {
			out.PassingCount++
		}
	}
	loops, problems := summarizeClosedLoops(m, out.Checks)
	out.ClosedLoops = loops
	out.Problems = append(out.Problems, problems...)
	out.ClosedVerified = countVerifiedClosed(loops)
	out.CheckCount = len(out.Checks)
	out.ProblemCount = len(out.Problems)
	out.VerifiedCount = countVerifiedEvidence(out.Checks, m.Required)
	if out.ProblemCount > 0 {
		out.Status = statusFailed
	}
	return out
}

func collectEvidence(out *report, dir string, item requiredEvidence) {
	path := filepath.Join(dir, item.File)
	data, err := readEvidence(path)
	if err != nil {
		out.Problems = append(out.Problems, missingEvidenceProblem(item, path, err))
		out.Checks = append(out.Checks, missingEvidenceCheck(item.ID))
		return
	}
	checks, problems := evaluate(item, data)
	out.Checks = append(out.Checks, checks...)
	out.Problems = append(out.Problems, problems...)
}

func missingEvidenceCheck(id string) checkSummary {
	return checkSummary{EvidenceID: id, Field: "__file__", Status: statusFailed}
}
