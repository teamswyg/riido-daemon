package main

func phaseCoverageRows(m manifest) []phaseCoverage {
	rows := make([]phaseCoverage, 0, len(m.RequiredPhases))
	for _, phase := range m.RequiredPhases {
		rows = append(rows, phaseCoverage{Phase: phase, Count: phaseCount(m.Loops, phase)})
	}
	return rows
}

func phaseCount(loops []loop, phase string) int {
	count := 0
	for _, item := range loops {
		if phasePresent(item, phase) {
			count++
		}
	}
	return count
}

func phasePresent(item loop, phase string) bool {
	switch phase {
	case "observe":
		return item.Observation.Summary != ""
	case "hypothesis":
		return item.Hypothesis.Summary != ""
	case "execute":
		return item.Execution.Summary != ""
	case "evaluate":
		return item.Evaluation.Summary != ""
	case "retrospective":
		return item.Retrospective.Summary != ""
	default:
		return false
	}
}
