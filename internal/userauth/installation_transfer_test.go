package userauth

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestTransferRecoveryNeverReturnsASigningProof(t *testing.T) {
	c := &Client{}
	c.config.Resource = "https://fixture.invalid"
	c.config.ClientID = "riido-cli"
	id := strings.Repeat("A", 43)
	q := DeviceLinkRequest{RequestID: id, ExpectedDeviceID: "new-device"}
	fingerprint := strings.Repeat("a", 64)
	p := TransferProof{DeviceLinkProof: DeviceLinkProof{Resource: c.config.Resource, ClientID: c.config.ClientID, InstallationID: "installation", IntentID: id, ChallengeHash: fingerprint, KeyFingerprint: fingerprint, ExpectedDeviceID: q.ExpectedDeviceID, ApprovedPrincipalID: "human", ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339Nano)}, PreviousDeviceID: "previous", PreviousLinkRevision: "1"}
	raw, _ := json.Marshal(TransferResult{RequestID: id, State: "proof_required", Proof: &p})
	if _, err := c.decodeTransfer(raw, q, fingerprint, true); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"recover": string(raw), "wrong-target": strings.Replace(string(raw), "new-device", "other", 1), "overflow": strings.Replace(string(raw), `"previous_link_revision":"1"`, `"previous_link_revision":"18446744073709551615"`, 1), "duplicate": strings.Replace(string(raw), `"state":`, `"state":"pending","state":`, 1), "receipt-with-proof": strings.Replace(string(raw), "proof_required", "complete", 1)} {
		t.Run(name, func(t *testing.T) {
			if _, err := c.decodeTransfer([]byte(body), q, fingerprint, name != "recover"); err == nil {
				t.Fatal("accepted unsafe proof or receipt")
			}
		})
	}
}
