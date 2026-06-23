package controlplane

import "context"

type claimLongPollKey struct{}

// ContextWithClaimLongPoll controls whether a ClaimTask call may wait for new
// work. Sources that do not support long-polling can ignore this hint.
func ContextWithClaimLongPoll(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, claimLongPollKey{}, enabled)
}

// ClaimLongPollEnabled reports the claim wait hint. The default is enabled so
// existing source implementations keep their previous behavior.
func ClaimLongPollEnabled(ctx context.Context) bool {
	enabled, ok := ctx.Value(claimLongPollKey{}).(bool)
	if !ok {
		return true
	}
	return enabled
}
