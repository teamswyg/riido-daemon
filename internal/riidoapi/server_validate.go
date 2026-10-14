package riidoapi

import (
	"context"
	"encoding/json"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/taskvalidation"
)

func (s Server) validateTask(ctx context.Context, params json.RawMessage) (ValidateResponse, error) {
	req, err := decodeValidateRequest(params)
	if err != nil {
		return ValidateResponse{}, err
	}
	db, err := taskdb.LoadTaskDB(s.config.TaskDBPath)
	if err != nil {
		return ValidateResponse{}, err
	}

	result, err := taskvalidation.Run(ctx, db, validationRunRequest(req), time.Now())
	if err != nil {
		return ValidateResponse{}, err
	}
	if err := taskdb.SaveTaskDB(s.config.TaskDBPath, result.TaskDB); err != nil {
		return ValidateResponse{}, err
	}
	return validateResponse(s.config.TaskDBPath, result), nil
}
