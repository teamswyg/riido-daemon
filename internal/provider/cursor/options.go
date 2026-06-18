package cursor

import "github.com/teamswyg/riido-daemon/internal/policy"

type StartOptions struct {
	Executable string
	// Profile picks the cursor-agent launch shape. Empty → DefaultProfile.
	Profile Profile
	// AllowYolo opts into Cursor's --yolo (auto-approve every tool).
	// Default false. Must be selected by the caller's security policy.
	AllowYolo bool
	// TrustTier and UnsafeBypassAllowed are consulted only when AllowYolo
	// is true. Host / Unknown deny regardless of UnsafeBypassAllowed.
	TrustTier           policy.TrustTier
	UnsafeBypassAllowed bool
}
