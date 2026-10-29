package saasplane

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

var errTerminationACK = errors.New("execution termination acknowledgement invalid")

// encoding/json accepts duplicate members and null numeric values. A receipt is
// an exact six-field object; reject those ambiguities before matching the claim.
func decodeTerminationReceipt(payload []byte) (terminationReceipt, error) {
	var ack terminationReceipt
	decoder := json.NewDecoder(bytes.NewReader(payload))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return ack, errTerminationACK
	}
	seen := make(map[string]bool, 6)
	for decoder.More() {
		token, err = decoder.Token()
		if err != nil {
			return ack, errTerminationACK
		}
		key, ok := token.(string)
		if !ok || seen[key] {
			return ack, errTerminationACK
		}
		switch key {
		case "schema_version", "kind", "assignment_id", "lease_epoch", "started_invocations", "recorded_at":
		default:
			return ack, errTerminationACK
		}
		seen[key] = true
		var value json.RawMessage
		if decoder.Decode(&value) != nil || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return ack, errTerminationACK
		}
	}
	token, err = decoder.Token()
	if err != nil || token != json.Delim('}') || len(seen) != 6 {
		return ack, errTerminationACK
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF || json.Unmarshal(payload, &ack) != nil {
		return ack, errTerminationACK
	}
	return ack, nil
}
