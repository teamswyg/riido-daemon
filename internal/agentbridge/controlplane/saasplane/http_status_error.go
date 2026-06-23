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
