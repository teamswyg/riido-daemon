package userapi

import (
	"encoding/json"
	"errors"
)

// readScopeIntegrityReason is the monolith fault reason for stored data that
// cannot be served safely (RIID-6008).
const readScopeIntegrityReason = "read_scope_integrity"

// Only stable, allowlisted server codes (and the integrity reason) are used;
// messages and other extensions never cross the public CLI boundary. Partial
// data remains a failed read.
func classifyGraphQLErrors(rawErrors []json.RawMessage) error {
	var classified error
	for _, raw := range rawErrors {
		var item struct {
			Extensions struct {
				Code   string `json:"code"`
				Reason string `json:"reason"`
			} `json:"extensions"`
		}
		if json.Unmarshal(raw, &item) != nil {
			return ErrGraphQL
		}
		var current error
		switch item.Extensions.Code {
		case "authentication_required":
			current = ErrAuthentication
		case "permission_denied":
			current = ErrPermission
		case "not_found":
			current = ErrNotFound
		case "conflict":
			current = ErrConflict
			if item.Extensions.Reason == readScopeIntegrityReason {
				current = ErrDataIntegrity
			}
		case "invalid_request":
			current = ErrInvalidRequest
		default:
			return ErrGraphQL
		}
		if classified != nil && !errors.Is(classified, current) {
			return ErrGraphQL
		}
		classified = current
	}
	if classified == nil {
		return ErrGraphQL
	}
	return classified
}
