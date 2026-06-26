package main

import (
	"encoding/json"
	"os"
)

type productRunOutcomeSource struct {
	State                string
	CoverageStatus       string
	DeploymentGateStatus string
	ScenarioStatus       map[string]string
}

type productRunOutcomeFile struct {
	CoverageStatus string `json:"coverage_status"`
	DeploymentGate struct {
		Status string `json:"status"`
	} `json:"deployment_gate"`
	Coverage struct {
		Rows []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"rows"`
	} `json:"coverage"`
}

func loadProductRunOutcome(root, rel string) productRunOutcomeSource {
	out := productRunOutcomeSource{State: localQARunMissing, ScenarioStatus: map[string]string{}}
	if rel == "" || !localQARunPresent(root, rel) {
		return out
	}
	data, err := os.ReadFile(repoPath(root, rel))
	if err != nil {
		out.State = localQARunInvalid
		return out
	}
	var file productRunOutcomeFile
	if err := json.Unmarshal(data, &file); err != nil {
		out.State = localQARunInvalid
		return out
	}
	out.State = localQARunFresh
	out.CoverageStatus = file.CoverageStatus
	out.DeploymentGateStatus = file.DeploymentGate.Status
	for _, row := range file.Coverage.Rows {
		out.ScenarioStatus[row.ID] = row.Status
	}
	return out
}

func outcomeEvidenceLinked(ids []string, run productRunOutcomeSource) bool {
	if run.State != localQARunFresh || len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		if !acceptedOutcomeStatus(run.ScenarioStatus[id]) {
			return false
		}
	}
	return true
}

func acceptedOutcomeStatus(status string) bool {
	return status == statusPassed || status == "observed"
}
