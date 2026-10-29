package standalonecli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func phaseFixture(t *testing.T, version, phase string) (*ProtectedStore, *protectedNativeFixture, LinkBinding) {
	t.Helper()
	s, n, c, b := newProtectedFixture(t)
	pending := map[string]any{"resource": b.Resource, "client_id": b.ClientID, "session_id": b.SessionID, "key_fingerprint": b.KeyFingerprint, "expected_device_id": b.ExpectedDeviceID, "request_id": strings.Repeat("A", 43), "recovery_secret": strings.Repeat("C", 42) + "A"}
	if phase != "" {
		pending["phase"] = phase
	}
	raw, err := json.Marshal(map[string]any{"schema_version": version, "credential": c, "pending_link": pending})
	if err != nil {
		t.Fatal(err)
	}
	n.raw = raw
	return s, n, b
}

func pendingPhase(t *testing.T, p PendingLink) string {
	t.Helper()
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if json.Unmarshal(raw, &m) != nil {
		t.Fatal("bad test pending")
	}
	result, _ := m["phase"].(string)
	return result
}

func TestProtectedLinkPhaseReadsV1WithoutRewritingAndMigratesOnSave(t *testing.T) {
	s, n, b := phaseFixture(t, "riido-cli-protected-session.v1", "")
	original := bytes.Clone(n.raw)
	p, err := s.LoadPendingLink(t.Context(), b.SessionID)
	if err != nil || p == nil || pendingPhase(t, *p) != "prepared" || !bytes.Equal(original, n.raw) || n.saves != 0 {
		t.Fatal("v1 pending was not read-only prepared", err)
	}
	credential, err := s.Load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if err = saveProtectedCredential(t, s, credential); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		SchemaVersion string         `json:"schema_version"`
		Pending       map[string]any `json:"pending_link"`
	}
	if json.Unmarshal(n.raw, &envelope) != nil || envelope.SchemaVersion != "riido-cli-protected-session.v2" || envelope.Pending["phase"] != "prepared" {
		t.Fatal("next successful save did not preserve phase in v2")
	}
}

func TestProtectedLinkPhaseV2AndCreatedACKPersistOriginalRequest(t *testing.T) {
	s, n, b := phaseFixture(t, "riido-cli-protected-session.v2", "prepared")
	p, err := s.LoadPendingLink(t.Context(), b.SessionID)
	if err != nil || p == nil {
		t.Fatal("prepared v2 could not resume", err)
	}
	created, err := s.MarkPendingLinkCreated(t.Context(), *p)
	if err != nil || pendingPhase(t, created) != "created" || created.RequestID != p.RequestID || created.RecoverySecret != p.RecoverySecret {
		t.Fatal("ACK phase lost original request", err)
	}
	saves := n.saves
	replay, err := s.MarkPendingLinkCreated(t.Context(), *p)
	if err != nil || replay != created || n.saves != saves {
		t.Fatal("ACK replay changed original phase", err)
	}
	restarted := &ProtectedStore{Native: n, Resource: s.Resource}
	got, err := restarted.LoadPendingLink(t.Context(), b.SessionID)
	if err != nil || got == nil || *got != created {
		t.Fatal("restart lost created phase", err)
	}
}

func TestProtectedLinkPhaseBeginPersistsPreparedV2(t *testing.T) {
	s, n, _, b := newProtectedFixture(t)
	p, err := s.BeginPendingLink(t.Context(), b)
	if err != nil || pendingPhase(t, p) != "prepared" {
		t.Fatal("new pending not prepared", err)
	}
	var envelope struct {
		SchemaVersion string `json:"schema_version"`
	}
	if json.Unmarshal(n.raw, &envelope) != nil || envelope.SchemaVersion != "riido-cli-protected-session.v2" {
		t.Fatal("new pending does not use phase-bearing version")
	}
}

func TestProtectedLinkPhaseACKSaveFailureKeepsOriginalCustody(t *testing.T) {
	for _, committed := range []bool{false, true} {
		t.Run(fmt.Sprint(committed), func(t *testing.T) {
			s, n, b := phaseFixture(t, "riido-cli-protected-session.v2", "prepared")
			p, err := s.LoadPendingLink(t.Context(), b.SessionID)
			if err != nil || p == nil {
				t.Fatal(err)
			}
			before := bytes.Clone(n.raw)
			n.saveErr, n.commitOnFail = securecredential.ErrUnavailable, committed
			got, err := s.MarkPendingLinkCreated(t.Context(), *p)
			if !errors.Is(err, securecredential.ErrUnavailable) || got != (PendingLink{}) || n.deletes != 0 {
				t.Fatal("failed ACK save exposed or deleted request", err)
			}
			if !committed && !bytes.Equal(before, n.raw) {
				t.Fatal("failed ACK write changed record")
			}
			n.saveErr = nil
			restarted := &ProtectedStore{Native: n, Resource: s.Resource}
			retry, err := restarted.MarkPendingLinkCreated(t.Context(), *p)
			if err != nil || retry.RequestID != p.RequestID || retry.RecoverySecret != p.RecoverySecret || pendingPhase(t, retry) != "created" {
				t.Fatal("ACK retry replaced pair", err)
			}
		})
	}
}

func TestProtectedLinkPhaseRejectsUnknownAndCrossVersionFieldsWithoutMutation(t *testing.T) {
	for _, kind := range []string{"missing", "unknown", "v1-phase", "duplicate", "null"} {
		t.Run(kind, func(t *testing.T) {
			version, phase := "riido-cli-protected-session.v2", "prepared"
			switch kind {
			case "missing":
				phase = ""
			case "unknown":
				phase = "approved"
			case "v1-phase":
				version = "riido-cli-protected-session.v1"
			}
			s, n, _ := phaseFixture(t, version, phase)
			if kind == "duplicate" {
				n.raw = bytes.Replace(n.raw, []byte(`"phase":"prepared"`), []byte(`"phase":"prepared","phase":"created"`), 1)
			}
			if kind == "null" {
				n.raw = bytes.Replace(n.raw, []byte(`"phase":"prepared"`), []byte(`"phase":null`), 1)
			}
			before := bytes.Clone(n.raw)
			if _, err := s.Load(t.Context()); err == nil {
				t.Fatal("unknown phase record loaded")
			}
			if err := saveProtectedCredential(t, s, userauth.Credential{SessionID: "replacement", AccessToken: "synthetic", TokenType: "Bearer"}); err == nil {
				t.Fatal("unknown phase overwritten")
			}
			if n.saves != 0 || n.deletes != 0 || !bytes.Equal(before, n.raw) {
				t.Fatal("invalid phase changed custody")
			}
		})
	}
}

func TestProtectedLinkPhaseACKCannotChangeAnotherRequestOrSession(t *testing.T) {
	s, n, b := phaseFixture(t, "riido-cli-protected-session.v2", "prepared")
	p, err := s.LoadPendingLink(t.Context(), b.SessionID)
	if err != nil || p == nil {
		t.Fatal(err)
	}
	before := bytes.Clone(n.raw)
	changed := *p
	changed.RequestID = strings.Repeat("B", 42) + "A"
	if _, err = s.MarkPendingLinkCreated(t.Context(), changed); err == nil || n.saves != 0 || !bytes.Equal(before, n.raw) {
		t.Fatal("foreign request advanced phase")
	}
	changed = *p
	changed.RecoverySecret = strings.Repeat("D", 42) + "A"
	if _, err = s.MarkPendingLinkCreated(t.Context(), changed); err == nil || n.saves != 0 || !bytes.Equal(before, n.raw) {
		t.Fatal("foreign secret advanced phase")
	}
	credential, err := s.Load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	credential.SessionID = "replacement-login"
	if err = saveProtectedCredential(t, s, credential); err != nil {
		t.Fatal(err)
	}
	replacement := bytes.Clone(n.raw)
	if _, err = s.MarkPendingLinkCreated(t.Context(), *p); !errors.Is(err, userauth.ErrLoggedOut) || !bytes.Equal(replacement, n.raw) {
		t.Fatal("old ACK restored earlier session", err)
	}
}

func TestProtectedLinkPhaseRefreshPreservesCreatedAndReplacementDropsIt(t *testing.T) {
	s, _, b := phaseFixture(t, "riido-cli-protected-session.v2", "created")
	original, err := s.LoadPendingLink(t.Context(), b.SessionID)
	if err != nil || original == nil {
		t.Fatal(err)
	}
	c, err := s.Load(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	c.AccessToken = "rotated-synthetic-access"
	if err = saveProtectedCredential(t, s, c); err != nil {
		t.Fatal(err)
	}
	got, err := s.LoadPendingLink(t.Context(), b.SessionID)
	if err != nil || got == nil || *got != *original {
		t.Fatal("same-session refresh lost ACK phase", err)
	}
	c.SessionID = "next-login"
	if err = saveProtectedCredential(t, s, c); err != nil {
		t.Fatal(err)
	}
	got, err = s.LoadPendingLink(t.Context(), c.SessionID)
	if err != nil || got != nil {
		t.Fatal("new login inherited original phase", err)
	}
}
