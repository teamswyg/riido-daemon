package claude

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/policy"
)

func validateStartPermission(opts StartOptions) error {
	if opts.PermissionMode == "" {
		return fmt.Errorf("%s: PermissionMode is required (no implicit bypass — see docs/20-domain/security.md)", Name)
	}
	if opts.PermissionMode != PermissionModeBypassDangerous {
		return nil
	}
	if opts.BetaFullAccessAllowed {
		return nil
	}
	decision := policy.EvaluateUnsafeBypass(policy.UnsafeBypassInput{
		TrustTier:    opts.TrustTier,
		Surface:      policy.UnsafeBypassClaudePermissions,
		BundleAllows: opts.UnsafeBypassAllowed,
	})
	if !decision.Allowed {
		return fmt.Errorf("%s: %s: %s", Name, decision.Code, decision.Reason)
	}
	return nil
}
