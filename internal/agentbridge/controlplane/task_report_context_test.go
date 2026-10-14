package controlplane

import (
	"context"
	"testing"
)

func TestTaskReportContextRoundTripFromMetadata(t *testing.T) {
	report, ok := TaskReportContextFromMetadata(map[string]string{
		MetadataRuntimeLeaseID:               "runtime-lease:t-1:3",
		MetadataRuntimeFencingToken:          "3",
		MetadataRuntimeCapabilityFingerprint: "fp-1",
	})
	if !ok {
		t.Fatal("TaskReportContextFromMetadata returned ok=false")
	}
	if report.RuntimeLeaseID != "runtime-lease:t-1:3" {
		t.Fatalf("runtime lease id: %q", report.RuntimeLeaseID)
	}
	if report.RuntimeFencingToken != 3 || !report.RuntimeFencingTokenSet {
		t.Fatalf("runtime fencing token: %+v", report)
	}
	if report.RuntimeCapabilityFingerprint != "fp-1" {
		t.Fatalf("runtime capability fingerprint: %+v", report)
	}

	ctx := ContextWithTaskReport(context.Background(), report)
	got, ok := TaskReportContextFromContext(ctx)
	if !ok || got != report {
		t.Fatalf("context round trip = %+v, %v", got, ok)
	}
}
