package main

import "github.com/teamswyg/riido-daemon/internal/scheduling"

type SurfaceEvidence struct {
	Name          string `json:"name"`
	MissingCode   string `json:"missing_code"`
	EligibleWhen  bool   `json:"eligible_when_supported"`
	CandidateFlag string `json:"candidate_flag"`
}

func validateSurfaces(rows []Surface) ([]problem, []SurfaceEvidence) {
	var problems []problem
	out := make([]SurfaceEvidence, 0, len(rows))
	seen := map[string]bool{}
	for _, row := range rows {
		rowProblems, evidence := validateSurface(row, seen)
		problems = append(problems, rowProblems...)
		out = append(out, evidence)
	}
	problems = append(problems, validateUnknownSurfaceFailsClosed()...)
	return problems, out
}

func validateSurface(row Surface, seen map[string]bool) ([]problem, SurfaceEvidence) {
	var problems []problem
	if seen[row.Name] {
		problems = append(problems, problem{Message: "duplicate surface " + row.Name})
	}
	seen[row.Name] = true
	if expected := expectedCandidateField(row.Name); expected != row.CandidateField {
		problems = append(problems, problem{Message: row.Name + " candidate field drift"})
	}
	if expected := expectedCapabilityFlag(row.Name); expected != row.CapabilityFlag {
		problems = append(problems, problem{Message: row.Name + " capability flag drift"})
	}
	if expected := expectedSchedulingConstant(row.Name); expected != row.SchedulingConstant {
		problems = append(problems, problem{Message: row.Name + " scheduling constant drift"})
	}
	missing := scheduling.EvaluateCapability(require(row.Name), candidateBase())
	supported, ok := candidateWithSurface(row.Name, true)
	if !ok {
		problems = append(problems, problem{Message: "unknown manifest surface " + row.Name})
	}
	eligible := scheduling.EvaluateCapability(require(row.Name), supported)
	problems = append(problems, requireMissingReason(row.Name, missing)...)
	problems = append(problems, requireEligible(row.Name, eligible)...)
	return problems, SurfaceEvidence{
		Name:          row.Name,
		MissingCode:   firstReasonCode(missing),
		EligibleWhen:  eligible.Eligible,
		CandidateFlag: row.CandidateField,
	}
}

func require(name string) scheduling.TaskRequirements {
	return scheduling.TaskRequirements{RequiredSurfaces: []scheduling.RequiredSurface{scheduling.RequiredSurface(name)}}
}
