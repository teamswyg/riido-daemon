package ingest

import "github.com/teamswyg/riido-contracts/ir"

type EnvelopeError struct {
	Violations []ir.EnvelopeViolation
}

func (e EnvelopeError) Error() string {
	if len(e.Violations) == 0 {
		return "ingest: invalid envelope"
	}
	first := e.Violations[0]
	if first.Field == "" {
		return "ingest: invalid envelope: " + first.Code
	}
	return "ingest: invalid envelope: " + first.Code + " " + first.Field
}
