package contextbroker

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// A content read asks for consecutive captured pages of one part so a large
// body is read in a few calls. The default matches the 48 KiB inline budget of
// work get; --budget changes it within the control plane's accepted range.
const (
	contentBatchDefaultBytes = 48 << 10
	contentBatchMinBytes     = 8 << 10
	contentBatchMaxBytes     = 256 << 10
)

func contentBatchBytes(args map[string]string) int {
	budget := budgetChars(args)
	if budget <= 0 {
		return contentBatchDefaultBytes
	}
	return min(max(budget, contentBatchMinBytes), contentBatchMaxBytes)
}

// fetchContent sends the read once. A control plane that predates batching
// rejects the unknown field while decoding, before it consumes or records a
// read, so only that refusal is resent as a single-page read.
func fetchContent(ctx context.Context, fetcher Fetcher, q Query) (Result, error) {
	result, err := fetcher.FetchAssignmentContext(ctx, q)
	if q.MaxBytes == 0 || !unsupportedContentBatch(err) {
		return result, err
	}
	q.MaxBytes = 0
	return fetcher.FetchAssignmentContext(ctx, q)
}

func unsupportedContentBatch(err error) bool {
	var fetchErr *FetchError
	return errors.As(err, &fetchErr) && fetchErr.Status == http.StatusBadRequest && strings.Contains(fetchErr.Message, `unknown field "max_bytes"`)
}
