// Package securecredential exchanges opaque credentials with a verified native
// helper. It never provides a plaintext or unprotected keychain fallback.
package securecredential

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
)

const (
	maxMessageBytes = 64 << 10
	maxValueBytes   = 32 << 10
)

var validKey = regexp.MustCompile(`^[a-zA-Z0-9_-]{16,128}$`)
var (
	ErrUnavailable         = errors.New("protected credential storage is unavailable")
	ErrInteractionRequired = errors.New("OS user confirmation is required")
	ErrCancelled           = errors.New("credential access was cancelled")
	ErrNotFound            = errors.New("no protected credential is stored")
)

type Store struct {
	HelperPath  string
	Key         string
	Interactive bool
	LockDir     string
	verify      func(context.Context, string) error
	run         func(context.Context, string, []byte) ([]byte, error)
}
type request struct {
	Version     int    `json:"version"`
	Operation   string `json:"operation"`
	Key         string `json:"key"`
	Interactive bool   `json:"interactive"`
	Value       string `json:"value,omitempty"`
}
type response struct {
	Version int  `json:"version"`
	OK      bool `json:"ok"`
	Data    *struct {
		Present   *bool   `json:"present"`
		Value     *string `json:"value"`
		Confirmed *bool   `json:"confirmed"`
	} `json:"data"`
	Error *struct {
		Code string `json:"code"`
	} `json:"error"`
}

func (s *Store) Load(ctx context.Context) ([]byte, error) {
	r, err := s.call(ctx, "load", nil)
	if err != nil {
		return nil, err
	}
	if !*r.Data.Present {
		return nil, ErrNotFound
	}
	value, err := base64.StdEncoding.DecodeString(*r.Data.Value)
	if err != nil || len(value) == 0 || len(value) > maxValueBytes {
		return nil, ErrUnavailable
	}
	return value, nil
}

func (s *Store) Save(ctx context.Context, value []byte) error {
	if len(value) == 0 || len(value) > maxValueBytes {
		return ErrUnavailable
	}
	_, err := s.call(ctx, "save", value)
	return err
}
func (s *Store) Delete(ctx context.Context) error { _, err := s.call(ctx, "delete", nil); return err }
func (s *Store) Status(ctx context.Context) (bool, error) {
	r, err := s.call(ctx, "status", nil)
	if err != nil {
		return false, err
	}
	return *r.Data.Present, nil
}

// ConfirmUserPresence is an explicit, fresh OS check before a new OAuth grant.
// Neither credential presence nor a successful storage operation is evidence.
func (s *Store) ConfirmUserPresence(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !s.Interactive {
		return ErrInteractionRequired
	}
	ctx, cancel := context.WithTimeout(ctx, helperTimeout)
	defer cancel()
	_, err := s.call(ctx, "confirm", nil)
	if canceled := ctx.Err(); canceled != nil {
		return canceled
	}
	return err
}

func (s *Store) call(ctx context.Context, op string, value []byte) (response, error) {
	var r response
	output, err := s.invoke(ctx, op, value)
	if err != nil {
		return r, err
	}
	if json.Unmarshal(output, &r) != nil || r.Version != 1 {
		return response{}, ErrUnavailable
	}
	if r.OK && r.Error == nil {
		if !validResponse(op, r) || op == "confirm" && !onlyConfirmationData(output) {
			return response{}, ErrUnavailable
		}
		return r, nil
	}
	if r.Error != nil {
		switch r.Error.Code {
		case "INTERACTION_REQUIRED":
			return r, ErrInteractionRequired
		case "CANCELLED":
			return r, ErrCancelled
		case "NOT_FOUND":
			return r, ErrNotFound
		case "TIMED_OUT":
			if op == "confirm" {
				return r, context.DeadlineExceeded
			}
		}
	}
	return r, ErrUnavailable
}

// invoke preserves the same verified, unreaped helper-process boundary for
// credential operations and typed installation proofs.
func (s *Store) invoke(ctx context.Context, op string, value []byte) ([]byte, error) {
	if !validKey.MatchString(s.Key) {
		return nil, ErrUnavailable
	}
	verify := s.verify
	if verify == nil {
		verify = verifyHelper
	}
	if err := verify(ctx, s.HelperPath); err != nil {
		return nil, ErrUnavailable
	}
	body, err := json.Marshal(request{Version: 1, Operation: op, Key: s.Key, Interactive: s.Interactive, Value: base64.StdEncoding.EncodeToString(value)})
	if err != nil {
		return nil, ErrUnavailable
	}
	run := s.run
	if run == nil {
		run = runHelper
	}
	output, err := run(ctx, s.HelperPath, body)
	if err != nil || len(output) > maxMessageBytes {
		return nil, ErrUnavailable
	}
	return output, nil
}

func onlyConfirmationData(output []byte) bool {
	var wire struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	return json.Unmarshal(output, &wire) == nil && len(wire.Data) == 1 && string(wire.Data["confirmed"]) == "true"
}

func validResponse(op string, r response) bool {
	if op == "confirm" {
		return r.Data != nil && r.Data.Confirmed != nil && *r.Data.Confirmed && r.Data.Present == nil && r.Data.Value == nil
	}
	if r.Data == nil || r.Data.Present == nil {
		return false
	}
	present := *r.Data.Present
	switch op {
	case "load":
		return present && r.Data.Value != nil && *r.Data.Value != "" || !present && r.Data.Value == nil
	case "save":
		return present && r.Data.Value == nil
	case "delete":
		return !present && r.Data.Value == nil
	case "status":
		return r.Data.Value == nil
	}
	return false
}
