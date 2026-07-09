package riidoapi

import (
	"encoding/json"
	"testing"
)

func TestDecodeResponseRejectsBadEnvelopeMethodAndData(t *testing.T) {
	var out map[string]string
	for name, body := range map[string][]byte{
		"bad json": []byte(`{`),
		"failure": mustMarshalResponse(t, responseEnvelope{
			Method: MethodStatus, OK: false, Error: "denied",
		}),
		"method mismatch": mustMarshalResponse(t, responseEnvelope{
			Method: MethodTasks, OK: true, Data: json.RawMessage(`{}`),
		}),
		"bad data": []byte(`{"ok":true,"method":"status","data":"oops"}`),
	} {
		if err := decodeResponse(body, string(MethodStatus), &out); err == nil {
			t.Fatalf("%s should fail", name)
		}
	}
}

func TestDecodeValidateRequestErrorsAndSuccess(t *testing.T) {
	for name, body := range map[string]json.RawMessage{
		"empty":        nil,
		"bad json":     json.RawMessage(`{`),
		"missing task": json.RawMessage(`{"command":"go test","approval_id":"appr"}`),
		"missing cmd":  json.RawMessage(`{"task_id":"task","approval_id":"appr"}`),
		"missing approval": json.RawMessage(
			`{"task_id":"task","command":"go test"}`),
		"negative timeout": json.RawMessage(
			`{"task_id":"task","command":"go test","approval_id":"appr","timeout_seconds":-1}`),
	} {
		if _, err := decodeValidateRequest(body); err == nil {
			t.Fatalf("%s should fail", name)
		}
	}
	req, err := decodeValidateRequest(json.RawMessage(
		`{"task_id":"task","command":"go test","approval_id":"appr"}`))
	if err != nil || req.TaskID != "task" {
		t.Fatalf("valid request failed: req=%#v err=%v", req, err)
	}
}

func TestSameStringsRequiresSameOrderAndLength(t *testing.T) {
	if !sameStrings([]string{"a", "b"}, []string{"a", "b"}) {
		t.Fatal("same strings should match")
	}
	if sameStrings([]string{"b", "a"}, []string{"a", "b"}) ||
		sameStrings([]string{"a"}, []string{"a", "b"}) {
		t.Fatal("different ordered or sized strings should not match")
	}
}

func mustMarshalResponse(t *testing.T, response responseEnvelope) []byte {
	t.Helper()
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
