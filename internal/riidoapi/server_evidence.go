package riidoapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func (s Server) addEvidence(params json.RawMessage) (EvidenceResponse, error) {
	var req EvidenceRequest
	if len(params) == 0 {
		return EvidenceResponse{}, errors.New("evidence params are required")
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return EvidenceResponse{}, fmt.Errorf("decode evidence params: %w", err)
	}
	db, err := taskdb.LoadTaskDB(s.config.TaskDBPath)
	if err != nil {
		return EvidenceResponse{}, err
	}
	updated, evidence, receipt, err := taskdb.AddGuardedTaskEvidence(db, evidenceInput(req), time.Now())
	if err != nil {
		return EvidenceResponse{}, err
	}
	if err := taskdb.SaveTaskDB(s.config.TaskDBPath, updated); err != nil {
		return EvidenceResponse{}, err
	}
	record, ok := findTask(updated, req.TaskID)
	if !ok {
		return EvidenceResponse{}, fmt.Errorf("task %s not found after evidence append", req.TaskID)
	}
	return EvidenceResponse{TaskDBPath: s.config.TaskDBPath, Task: record, Evidence: evidence, Receipt: receipt}, nil
}
