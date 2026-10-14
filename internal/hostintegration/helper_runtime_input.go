package hostintegration

// HelperRuntimePlanInput is reduced by C11 adapters before they install or
// start any helper process. It is pure data and does not call OS APIs.
type HelperRuntimePlanInput struct {
	Channel             DistributionChannel
	HostOS              HostOS
	AppDataRoot         AppDataRoot
	Consent             ConsentState
	StoreReviewApproved bool
	EndpointName        string
}
