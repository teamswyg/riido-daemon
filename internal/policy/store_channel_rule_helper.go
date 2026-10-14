package policy

func evaluateBackgroundHelper(input StoreChannelPolicyInput) Decision {
	if !input.ExplicitConsentGranted {
		return deny("STORE_CHANNEL_REQUIRES_CONSENT", "background helper requires explicit user consent")
	}
	if input.Channel.StoreManaged() && !input.StoreReviewApproved {
		return deny("STORE_CHANNEL_REQUIRES_STORE_REVIEW_APPROVAL", "store-managed background helper requires store policy review approval")
	}
	return allowStoreChannelSurface("background helper is allowed for this distribution channel")
}

func allowStoreChannelSurface(reason string) Decision {
	return Decision{Allowed: true, Code: "ALLOWED", Reason: reason}
}
