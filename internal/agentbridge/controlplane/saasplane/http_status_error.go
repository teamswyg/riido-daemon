package saasplane

import (
	"encoding/hex"
	"fmt"
)

type httpStatusError struct {
	RequestID  string
	Path       string
	Status     string
	StatusCode int
	Body       string
}

func (e httpStatusError) Error() string {
	return fmt.Sprintf("saasplane: %s returned %s: %s", e.Path, e.Status, e.Body)
}

// HTTPStatusCode exposes only the safe status for operational diagnostics.
func (e httpStatusError) HTTPStatusCode() int { return e.StatusCode }

// RequestCorrelationID excludes untrusted response text from operational logs.
func (e httpStatusError) RequestCorrelationID() string {
	if len(e.RequestID) != 32 {
		return ""
	}
	if _, err := hex.DecodeString(e.RequestID); err != nil {
		return ""
	}
	return e.RequestID
}
