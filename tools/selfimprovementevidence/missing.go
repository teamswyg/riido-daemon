package main

import "fmt"

func missingEvidenceProblem(item requiredEvidence, path string, err error) string {
	return fmt.Sprintf(
		"missing evidence %s: expected %s; produce with `%s`; cause: %v",
		item.ID,
		path,
		item.Producer,
		err,
	)
}
