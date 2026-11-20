package saasplane

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestTransferConfirmUsesNextRevisionAndNativeIdentity(t *testing.T) {
	for _, revision := range []string{"8", "1", "7", "9"} {
		t.Run(revision, func(t *testing.T) {
			id := strings.Repeat("A", 43)
			client, p := cliLinkTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/daemon/cli-installation-links/"+id+"/confirm" || r.Header.Get("X-Riido-Device-Secret") != "synthetic-device-secret" || r.Header.Get("Authorization") != "" {
					t.Error("wrong native authority")
				}
				ack := cliLinkTestACK(cliLinkTestProof())
				ack["link_revision"] = revision
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(ack)
			}))
			p.IntentID = id
			got, err := client.ConfirmTransfer(t.Context(), p, "source", "7", "MAYCAQECAQE")
			if (err == nil) != (revision == "8") || got.Confirmed != (revision == "8") {
				t.Fatalf("incorrect receipt for revision %s: %v", revision, err)
			}
		})
	}
}
