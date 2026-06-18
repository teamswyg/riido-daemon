package riidoapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func decodeValidateRequest(params json.RawMessage) (ValidateRequest, error) {
	var req ValidateRequest
	if len(params) == 0 {
		return ValidateRequest{}, errors.New("validate params are required")
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return ValidateRequest{}, fmt.Errorf("decode validate params: %w", err)
	}
	return req, validateRequest(req)
}

func validateRequest(req ValidateRequest) error {
	if strings.TrimSpace(req.TaskID) == "" {
		return errors.New("task_id is required")
	}
	if strings.TrimSpace(req.Command) == "" {
		return errors.New("command is required")
	}
	if strings.TrimSpace(req.ApprovalID) == "" {
		return errors.New("approval_id is required before validation command execution")
	}
	if req.TimeoutSeconds < 0 {
		return errors.New("timeout_seconds must not be negative")
	}
	return nil
}
