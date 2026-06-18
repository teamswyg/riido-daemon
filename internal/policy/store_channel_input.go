package policy

import "github.com/teamswyg/riido-daemon/internal/hostintegration"

// StoreChannelPolicyInput asks whether a distribution channel may use one
// concrete host surface. OSGrantPresent means the caller has already reduced
// platform-specific proof such as a security-scoped bookmark or package-local
// grant into a boolean fact.
type StoreChannelPolicyInput struct {
	Channel                hostintegration.DistributionChannel
	Surface                StoreSurface
	ExplicitConsentGranted bool
	OSGrantPresent         bool
	StoreReviewApproved    bool
}
