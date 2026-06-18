package agentbridge

import "testing"

func TestInjectTelemetryContractIsIdempotent(t *testing.T) {
	first := InjectTelemetryContract("do it")
	second := InjectTelemetryContract(first)
	if second != first {
		t.Fatalf("telemetry contract duplicated:\nfirst=%q\nsecond=%q", first, second)
	}
}
