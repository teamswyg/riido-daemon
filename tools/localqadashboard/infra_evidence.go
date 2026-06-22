package main

import (
	"encoding/json"
	"fmt"
)

type infraEvidence struct {
	TerraformManaged    bool   `json:"terraform_managed"`
	PublicAccessBlocked bool   `json:"public_access_blocked"`
	EncryptionAlgorithm string `json:"encryption_algorithm"`
	LifecycleExpireDays int    `json:"lifecycle_expire_days"`
	LatestIndexObserved bool   `json:"latest_index_observed"`
	LatestIndexBytes    int64  `json:"latest_index_bytes"`
	LatestCacheControl  string `json:"latest_cache_control"`
	LatestObjectSSE     string `json:"latest_object_sse"`
}

func infraEvidenceScenarios(path string) []externalScenario {
	data, ok := readOptional(path)
	if !ok {
		return nil
	}
	var evidence infraEvidence
	if json.Unmarshal(data, &evidence) != nil {
		return nil
	}
	status := statusPassed
	if !infraEvidencePassed(evidence) {
		status = statusFailed
	}
	return []externalScenario{{
		ID:             "infra.local_qa_dashboard",
		Status:         status,
		FailureSummary: infraEvidenceDetail(evidence),
	}}
}

func infraEvidencePassed(e infraEvidence) bool {
	return e.TerraformManaged && e.PublicAccessBlocked &&
		e.EncryptionAlgorithm == "AES256" && e.LifecycleExpireDays > 0 &&
		e.LatestIndexObserved && e.LatestIndexBytes > 0 &&
		e.LatestCacheControl == "no-store" && e.LatestObjectSSE == "AES256"
}

func infraEvidenceDetail(e infraEvidence) string {
	return fmt.Sprintf(
		"terraform_managed=%t public_access_blocked=%t encryption=%s lifecycle_days=%d latest_index=%t latest_bytes=%d cache_control=%s object_sse=%s",
		e.TerraformManaged, e.PublicAccessBlocked, e.EncryptionAlgorithm,
		e.LifecycleExpireDays, e.LatestIndexObserved, e.LatestIndexBytes,
		e.LatestCacheControl, e.LatestObjectSSE,
	)
}
