package main

import "github.com/teamswyg/riido-daemon/internal/scheduling"

type GateEvidence struct {
	Order int    `json:"order"`
	Code  string `json:"code"`
	Seen  bool   `json:"seen"`
}

func validateGates(gates []Gate) ([]problem, []GateEvidence) {
	var problems []problem
	out := make([]GateEvidence, 0, len(gates))
	for i, gate := range gates {
		if gate.Order != i+1 {
			problems = append(problems, problem{Message: gate.Code + " order drift"})
		}
		req, candidate, ok := gateScenario(gate.Code)
		if !ok {
			problems = append(problems, problem{Message: "unknown gate code " + gate.Code})
			continue
		}
		eval := scheduling.EvaluateCapability(req, candidate)
		seen := hasReason(eval, gate.Code)
		if !seen {
			problems = append(problems, problem{Message: gate.Code + " did not evaluate"})
		}
		out = append(out, GateEvidence{Order: gate.Order, Code: gate.Code, Seen: seen})
	}
	return problems, out
}

func hasReason(eval scheduling.Eligibility, code string) bool {
	for _, reason := range eval.Reasons {
		if reason.Code == code {
			return true
		}
	}
	return false
}
