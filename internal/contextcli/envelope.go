package contextcli

import (
	"crypto/rand"
	"encoding/hex"

	contract "github.com/teamswyg/riido-contracts/contextcli"
)

// The wire contract (envelope, sources, warnings, page, error) is owned by
// github.com/teamswyg/riido-contracts/contextcli; these aliases keep daemon
// call sites stable while the shape stays defined in one place.
const EnvelopeSchemaVersion = contract.EnvelopeSchemaVersion

type (
	Envelope     = contract.Envelope
	SourceStatus = contract.SourceStatus
	Source       = contract.Source
	Warning      = contract.Warning
	Page         = contract.Page
	Error        = contract.Error
)

const (
	SourceAvailable   = contract.SourceAvailable
	SourcePartial     = contract.SourcePartial
	SourceDenied      = contract.SourceDenied
	SourceUnavailable = contract.SourceUnavailable
	SourceUnsupported = contract.SourceUnsupported
	SourceDeleted     = contract.SourceDeleted
)

// NewEnvelope returns a success envelope for data.
func NewEnvelope(requestID string, data any) Envelope {
	return Envelope{
		SchemaVersion: EnvelopeSchemaVersion,
		RequestID:     requestID,
		Data:          data,
		Sources:       []Source{},
		Warnings:      []Warning{},
	}
}

// FailureEnvelope returns an envelope that carries only err.
func FailureEnvelope(requestID string, err *CLIError) Envelope {
	env := NewEnvelope(requestID, nil)
	env.Error = &Error{
		Code:       err.Code,
		Message:    err.Message,
		Retryable:  err.Retryable(),
		NextAction: err.NextAction,
		RequestID:  requestID,
	}
	return env
}

// NewRequestID returns a random identifier that ties CLI output to logs.
func NewRequestID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "cli_unknown"
	}
	return "cli_" + hex.EncodeToString(buf[:])
}
