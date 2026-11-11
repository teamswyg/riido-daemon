package saasplane

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/securecredential"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

type DeviceBootstrapSigner interface {
	ExistingInstallationKey(context.Context, string) (securecredential.InstallationPublicKey, error)
	SignDeviceBootstrapProof(context.Context, securecredential.InstallationPublicKey, securecredential.DeviceBootstrapProof) (securecredential.DeviceBootstrapSignature, error)
}

type BootstrapAdmission func(context.Context, userauth.BootstrapRequest) (userauth.DeviceLinkApproval, error)

// The private daemon process is the recipient of the native device credential.
// It obtains the challenge directly from its pinned CP, never from CLI input.
func DeviceBootstrap(ctx context.Context, action string, input []byte, signer DeviceBootstrapSigner, admission BootstrapAdmission) (string, error) {
	q, err := parseBootstrapInput(input, action)
	if err != nil {
		return "", err
	}
	return q.advance(ctx, action, signer, http.DefaultClient, admission)
}

func (q bootstrapRequest) advance(ctx context.Context, action string, signer DeviceBootstrapSigner, client *http.Client, admission BootstrapAdmission) (string, error) {
	if action != "status" && action != "prepare" && action != "activate" {
		return "", cliLinkFailure("BOOTSTRAP_INVALID")
	}
	call := func(op string, payload map[string]string) (map[string]string, error) {
		return q.post(ctx, client, op, payload)
	}
	current, err := call("status", nil)
	if err != nil {
		return "", err
	}
	state := current["state"]
	if action == "status" || state == "complete" || action == "prepare" && state == "prepared" {
		return state, nil
	}
	if action == "activate" && state != "prepared" {
		return "", cliLinkFailure("BOOTSTRAP_NOT_PREPARED")
	}
	if action == "prepare" {
		until, err := q.confirmAdmission(ctx, admission)
		if err != nil {
			return "", err
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, until)
		defer cancel()
	}
	expires, _ := bootstrapTimestamp(current["expires_at"])
	if !expires.After(time.Now().Add(5 * time.Second)) {
		// Prepared recovery survives an OS restart. Initial issuance can
		// renew only within the independently verified human approval above.
		refreshed, e := call("renew", nil)
		if e != nil {
			return "", e
		}
		if refreshed["state"] != state || refreshed["previous_boot_digest"] != current["previous_boot_digest"] {
			return "", cliLinkFailure("BOOTSTRAP_CHANGED")
		}
		current = refreshed
	}
	if signer == nil {
		return "", cliLinkFailure("BOOTSTRAP_KEY_UNAVAILABLE")
	}
	key, err := signer.ExistingInstallationKey(ctx, q.resource)
	if err != nil || key.KeyFingerprint != q.key {
		return "", cliLinkFailure("BOOTSTRAP_KEY_UNAVAILABLE")
	}
	proof := securecredential.DeviceBootstrapProof{Phase: action, TransitionID: q.id, PreviousBootDigest: current["previous_boot_digest"], DeviceLinkProof: securecredential.DeviceLinkProof{Resource: q.resource, ClientID: "riido-cli", InstallationID: q.installation, IntentID: current["intent_id"], ChallengeHash: current["challenge_digest"], KeyFingerprint: q.key, ExpectedDeviceID: q.device, ApprovedPrincipalID: q.owner, ExpiresAt: current["expires_at"]}}
	signature, e := signer.SignDeviceBootstrapProof(ctx, key, proof)
	if e != nil {
		return "", cliLinkFailure("BOOTSTRAP_WITNESS_UNAVAILABLE")
	}
	if !cliLinkDigest(signature.BootDigest) || !cliLinkSignature(signature.Signature) || action == "activate" && signature.BootDigest == proof.PreviousBootDigest {
		return "", cliLinkFailure("BOOTSTRAP_INVALID_PROOF")
	}
	if action == "prepare" {
		// OS user presence may outlive the original user, grant, or consent.
		if _, err := q.confirmAdmission(ctx, admission); err != nil {
			return "", err
		}
	}
	result, e := call(action, map[string]string{"intent_id": proof.IntentID, "signature": signature.Signature, "boot_digest": signature.BootDigest})
	if e != nil {
		return "", e
	}
	if action == "prepare" && result["state"] == "prepared" && result["previous_boot_digest"] == signature.BootDigest {
		return "prepared", nil
	}
	if action == "activate" && result["state"] == "complete" && result["previous_boot_digest"] == proof.PreviousBootDigest {
		return "complete", nil
	}
	return "", cliLinkFailure("BOOTSTRAP_INVALID_RESPONSE")
}

func (q bootstrapRequest) confirmAdmission(ctx context.Context, admission BootstrapAdmission) (time.Time, error) {
	if admission == nil {
		return time.Time{}, cliLinkFailure("BOOTSTRAP_APPROVAL_UNAVAILABLE")
	}
	result, err := admission(ctx, userauth.BootstrapRequest{
		BootstrapIdentity: userauth.BootstrapIdentity{Resource: q.resource, ClientID: "riido-cli", InstallationID: q.installation, KeyFingerprint: q.key, OwnerID: q.owner},
		RequestID:         q.id, DeviceID: q.device, NextHash: q.nextHash, RecoverySecret: q.recovery,
	})
	until := time.Unix(result.ExpiresAt, 0)
	if err != nil || result.RequestID != q.id || result.State != "issued" || result.ExpiresAt != q.approvalExpiry || !until.After(time.Now()) || ctx.Err() != nil {
		return time.Time{}, cliLinkFailure("BOOTSTRAP_APPROVAL_UNAVAILABLE")
	}
	return until, nil
}

func (q bootstrapRequest) post(ctx context.Context, transport *http.Client, op string, body map[string]string) (map[string]string, error) {
	if body == nil {
		body = map[string]string{}
	}
	body["schema_version"] = deviceBootstrapSchema
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, cliLinkFailure("BOOTSTRAP_INVALID")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, q.origin+"/v1/daemon/device-bootstraps/"+q.id+"/"+op, bytes.NewReader(raw))
	if err != nil {
		return nil, cliLinkFailure("BOOTSTRAP_INVALID")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Riido-Device-Id", q.device)
	req.Header.Set("X-Riido-Device-Secret", q.secret)
	client := *transport
	client.Jar = nil
	client.Timeout = 10 * time.Second
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	res, err := client.Do(req)
	if err != nil {
		return nil, cliLinkFailure("BOOTSTRAP_UNAVAILABLE")
	}
	defer res.Body.Close()
	media, _, e := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if res.StatusCode != http.StatusOK || e != nil || media != "application/json" {
		return nil, cliLinkFailure("BOOTSTRAP_UNAVAILABLE")
	}
	raw, err = io.ReadAll(io.LimitReader(res.Body, 8193))
	if err != nil {
		return nil, cliLinkFailure("BOOTSTRAP_INVALID_RESPONSE")
	}
	var phase struct {
		State string `json:"state"`
	}
	if json.Unmarshal(raw, &phase) != nil {
		return nil, cliLinkFailure("BOOTSTRAP_INVALID_RESPONSE")
	}
	keys := []string{"schema_version", "resource", "client_id", "installation_id", "key_fingerprint", "transition_id", "device_id", "owner_principal_id", "state", "intent_id", "challenge_digest", "expires_at"}
	switch phase.State {
	case "issued":
	case "prepared":
		keys = append(keys, "previous_boot_digest")
	case "complete":
		keys = append(keys, "previous_boot_digest", "execution_coverage_generation", "credential_generation", "credential_issued_at", "completed_at")
	default:
		return nil, cliLinkFailure("BOOTSTRAP_INVALID_RESPONSE")
	}
	f, e := bootstrapFields(raw, keys...)
	if e != nil || f["schema_version"] != deviceBootstrapSchema || f["resource"] != q.resource || f["client_id"] != "riido-cli" || f["installation_id"] != q.installation || f["key_fingerprint"] != q.key || f["transition_id"] != q.id || f["device_id"] != q.device || f["owner_principal_id"] != q.owner || !bootstrapRandomID(f["intent_id"]) || !cliLinkDigest(f["challenge_digest"]) {
		return nil, cliLinkFailure("BOOTSTRAP_INVALID_RESPONSE")
	}
	if _, ok := bootstrapTimestamp(f["expires_at"]); !ok {
		return nil, cliLinkFailure("BOOTSTRAP_INVALID_RESPONSE")
	}
	if phase.State != "issued" && !cliLinkDigest(f["previous_boot_digest"]) {
		return nil, cliLinkFailure("BOOTSTRAP_INVALID_RESPONSE")
	}
	if phase.State == "complete" {
		hash := sha256.Sum256([]byte(strings.Join([]string{deviceBootstrapSchema, q.id, q.device, q.key, q.nextHash}, "\x00")))
		if f["execution_coverage_generation"] != "covb_"+hex.EncodeToString(hash[:]) || !strings.HasPrefix(f["credential_generation"], "cred_") || !cliLinkDigest(strings.TrimPrefix(f["credential_generation"], "cred_")) {
			return nil, cliLinkFailure("BOOTSTRAP_INVALID_RESPONSE")
		}
		for _, k := range []string{"credential_issued_at", "completed_at"} {
			if _, ok := bootstrapTimestamp(f[k]); !ok {
				return nil, cliLinkFailure("BOOTSTRAP_INVALID_RESPONSE")
			}
		}
	}
	return f, nil
}
