package taskvalidation

import (
	"context"
	"time"
)

func normalizeRunContext(ctx context.Context, now time.Time) (context.Context, time.Time) {
	if now.IsZero() {
		now = time.Now()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx, now
}
