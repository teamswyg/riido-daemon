package cursor

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func appendUnsafeBypass(args []string, opts StartOptions) ([]string, error) {
	if !opts.AllowYolo {
		return args, nil
	}
	decision := policy.EvaluateUnsafeBypass(policy.UnsafeBypassInput{
		TrustTier:    opts.TrustTier,
		Surface:      policy.UnsafeBypassCursorYolo,
		BundleAllows: opts.UnsafeBypassAllowed,
	})
	if !decision.Allowed {
		return nil, fmt.Errorf("cursor: %s: %s", decision.Code, decision.Reason)
	}
	return append(args, "--yolo"), nil
}
