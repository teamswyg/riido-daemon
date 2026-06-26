package main

import (
	"encoding/json"
	"os"
	"time"
)

const (
	localQARunMissing = "missing"
	localQARunFresh   = "fresh"
	localQARunExpired = "expired"
	localQARunInvalid = "invalid"
)

type localQARunFile struct {
	ExpiresAt string `json:"expires_at"`
}

func localQARunPresent(root, rel string) bool {
	if rel == "" {
		return false
	}
	_, err := os.Stat(repoPath(root, rel))
	return err == nil
}

func localQARunFreshness(root, rel string, now time.Time) (string, string) {
	if !localQARunPresent(root, rel) {
		return localQARunMissing, ""
	}
	data, err := os.ReadFile(repoPath(root, rel))
	if err != nil {
		return localQARunInvalid, ""
	}
	var file localQARunFile
	if err := json.Unmarshal(data, &file); err != nil || file.ExpiresAt == "" {
		return localQARunInvalid, file.ExpiresAt
	}
	expires, err := time.Parse(time.RFC3339, file.ExpiresAt)
	if err != nil {
		return localQARunInvalid, file.ExpiresAt
	}
	if now.Before(expires) {
		return localQARunFresh, file.ExpiresAt
	}
	return localQARunExpired, file.ExpiresAt
}
