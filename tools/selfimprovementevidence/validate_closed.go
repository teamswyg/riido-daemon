package main

import (
	"errors"
	"fmt"
)

func validateClosedLoops(loops []closedLoop, evidence map[string]bool) error {
	kinds := map[string]bool{}
	seen := map[string]bool{}
	for _, item := range loops {
		if item.ID == "" || item.Kind == "" || item.Description == "" {
			return errors.New("closed loop id, kind, and description are required")
		}
		if seen[item.ID] {
			return fmt.Errorf("duplicate closed loop %s", item.ID)
		}
		seen[item.ID] = true
		kinds[item.Kind] = true
		if err := validateClosedLoopEvidence(item, evidence); err != nil {
			return err
		}
	}
	if !kinds["bug"] || !kinds["feature"] {
		return errors.New("closed loops must include bug and feature kinds")
	}
	return nil
}

func validateClosedLoopEvidence(item closedLoop, evidence map[string]bool) error {
	if len(item.EvidenceIDs) == 0 {
		return fmt.Errorf("%s has no evidence_ids", item.ID)
	}
	for _, id := range item.EvidenceIDs {
		if !evidence[id] {
			return fmt.Errorf("%s references unknown evidence %s", item.ID, id)
		}
	}
	return nil
}
