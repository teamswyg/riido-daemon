package saasplane

import "fmt"

type httpStatusError struct {
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
