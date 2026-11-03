package controlplane

import "context"

type claimCapacityKey struct{}

// Notification sources retain hints while local preparation/execution/delivery
// occupies capacity. Existing polling adapters may ignore this optional hint.
func ContextWithClaimCapacity(ctx context.Context, available bool) context.Context {
	return context.WithValue(ctx, claimCapacityKey{}, available)
}

func ClaimCapacityAvailable(ctx context.Context) bool {
	available, present := ctx.Value(claimCapacityKey{}).(bool)
	return !present || available
}
