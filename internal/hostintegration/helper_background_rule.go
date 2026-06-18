package hostintegration

// HelperBackgroundRule records which external approval facts are required
// before the helper may run after the foreground app exits.
type HelperBackgroundRule string

const (
	HelperBackgroundExplicitConsent       HelperBackgroundRule = "explicit-consent"
	HelperBackgroundConsentAndStoreReview HelperBackgroundRule = "explicit-consent-and-store-review"
)
