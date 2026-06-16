package controlplane

import (
	"context"
	"strconv"
	"strings"
)

// TaskReportContext carries claim-time lease metadata alongside reporter
// calls without widening every reporter method signature.
type TaskReportContext struct {
	RuntimeLeaseID               string
	RuntimeFencingToken          int64
	RuntimeFencingTokenSet       bool
	RuntimeCapabilityFingerprint string
}

type taskReportContextKey struct{}

// ContextWithTaskReport attaches claim-time lease metadata to reporter calls.
func ContextWithTaskReport(ctx context.Context, report TaskReportContext) context.Context {
	return context.WithValue(ctx, taskReportContextKey{}, report)
}

// TaskReportContextFromContext returns claim-time lease metadata attached to ctx.
func TaskReportContextFromContext(ctx context.Context) (TaskReportContext, bool) {
	report, ok := ctx.Value(taskReportContextKey{}).(TaskReportContext)
	return report, ok
}

// TaskReportContextFromMetadata extracts claim-time lease metadata from a task request.
func TaskReportContextFromMetadata(metadata map[string]string) (TaskReportContext, bool) {
	if len(metadata) == 0 {
		return TaskReportContext{}, false
	}
	report := TaskReportContext{
		RuntimeLeaseID:               strings.TrimSpace(metadata[MetadataRuntimeLeaseID]),
		RuntimeCapabilityFingerprint: strings.TrimSpace(metadata[MetadataRuntimeCapabilityFingerprint]),
	}
	if raw := strings.TrimSpace(metadata[MetadataRuntimeFencingToken]); raw != "" {
		token, err := strconv.ParseInt(raw, 10, 64)
		if err == nil {
			report.RuntimeFencingToken = token
			report.RuntimeFencingTokenSet = true
		}
	}
	return report, report.RuntimeLeaseID != "" || report.RuntimeFencingTokenSet || report.RuntimeCapabilityFingerprint != ""
}
