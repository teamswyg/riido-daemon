package policy

// TrustTier describes the runtime isolation level from
// docs/20-domain/security.md §1.
type TrustTier string

const (
	TrustTierHost               TrustTier = "Host"
	TrustTierIsolatedContainer  TrustTier = "IsolatedContainer"
	TrustTierEphemeralVM        TrustTier = "EphemeralVM"
	TrustTierCIControlledRunner TrustTier = "CIControlledRunner"
	TrustTierUnknown            TrustTier = "Unknown"
)
