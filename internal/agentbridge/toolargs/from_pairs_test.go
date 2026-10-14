package toolargs

import "testing"

func TestToolArgsFromPairsRedactsSensitiveKeys(t *testing.T) {
	args := FromPairs("command", "go test ./...", "api_token", "secret-value")

	if args["command"] != "go test ./..." {
		t.Fatalf("command arg = %q", args["command"])
	}
	if args["api_token"] != RedactedValue {
		t.Fatalf("api token must be redacted: %+v", args)
	}
}
