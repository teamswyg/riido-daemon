package saasplane

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/executionproof"
)

const terminationSchema = "riido-execution-termination.v1"

var errTerminationUnsupported = errors.New("execution termination observation is unsupported")

type terminationRequest struct {
	SchemaVersion      string `json:"schema_version"`
	Kind               string `json:"kind"`
	StartedInvocations uint64 `json:"started_invocations"`
}

type terminationReceipt struct {
	SchemaVersion      string    `json:"schema_version"`
	Kind               string    `json:"kind"`
	AssignmentID       string    `json:"assignment_id"`
	LeaseEpoch         string    `json:"lease_epoch"`
	StartedInvocations uint64    `json:"started_invocations"`
	RecordedAt         time.Time `json:"recorded_at"`
}

func validTerminationHeader(value string, limit int) bool {
	if len(value) == 0 || len(value) > limit {
		return false
	}
	for _, b := range []byte(value) {
		if b < 33 || b > 126 {
			return false
		}
	}
	return true
}

// This endpoint deliberately does not use generic doJSON: errors and redirects
// must never reflect or forward its protected lease/device credentials.
func (p *Plane) postExecutionTermination(ctx context.Context, b *executionClaimBinding, proof executionproof.Proof) error {
	if p.cfg.DeviceID != b.deviceID || !validTerminationHeader(b.deviceID, 512) || !validTerminationHeader(p.cfg.DeviceSecret, 8192) || !validTerminationHeader(b.assignment.LeaseToken, 8192) || proof.Kind != executionproof.QuiescentKind {
		return errors.New("execution termination credentials unavailable")
	}
	data, err := json.Marshal(terminationRequest{SchemaVersion: terminationSchema, Kind: proof.Kind, StartedInvocations: proof.StartedInvocations})
	if err != nil {
		return errors.New("execution termination request invalid")
	}
	ctx, cancel := context.WithTimeout(ctx, p.cfg.RequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.BaseURL+"/v1/daemon/execution-claims/"+url.PathEscape(b.assignment.ID)+"/termination", bytes.NewReader(data))
	if err != nil {
		return errors.New("execution termination request invalid")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Riido-Device-Id", b.deviceID)
	req.Header.Set("X-Riido-Device-Secret", p.cfg.DeviceSecret)
	req.Header.Set("X-Riido-Assignment-Lease-Token", b.assignment.LeaseToken)
	client := *p.client
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("execution termination transport failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
		return errTerminationUnsupported
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("execution termination rejected (HTTP %d)", resp.StatusCode)
	}
	payload, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if err != nil || len(payload) > 4096 {
		return errors.New("execution termination acknowledgement invalid")
	}
	ack, err := decodeTerminationReceipt(payload)
	if err != nil {
		return err
	}
	digest := sha256.Sum256([]byte(b.assignment.LeaseToken))
	if ack.SchemaVersion != terminationSchema || ack.Kind != proof.Kind || ack.AssignmentID != b.assignment.ID || ack.LeaseEpoch != hex.EncodeToString(digest[:]) || ack.StartedInvocations != proof.StartedInvocations || ack.RecordedAt.IsZero() {
		return errors.New("execution termination acknowledgement mismatch")
	}
	return nil
}
