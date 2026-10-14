package hostintegration

import (
	"fmt"
)

// ResolveHelperRuntimePlan returns the helper/tray role and local-only IPC
// contract for a distribution channel. It does not spawn providers, install
// startup entries, open sockets/pipes, or inspect platform entitlements.
func ResolveHelperRuntimePlan(in HelperRuntimePlanInput) (HelperRuntimePlan, error) {
	if !in.Channel.Valid() {
		return HelperRuntimePlan{}, fmt.Errorf("unknown distribution channel %q", in.Channel)
	}
	if !in.HostOS.Valid() {
		return HelperRuntimePlan{}, fmt.Errorf("unknown host OS %q", in.HostOS)
	}
	if err := validateHelperRuntimeAppDataRoot(in.Channel, in.AppDataRoot); err != nil {
		return HelperRuntimePlan{}, err
	}

	endpoint, err := helperRuntimeEndpoint(in)
	if err != nil {
		return HelperRuntimePlan{}, err
	}

	plan := baseHelperRuntimePlan(in, endpoint)
	if !applyHelperRuntimeChannelPlan(&plan, in) {
		return HelperRuntimePlan{}, fmt.Errorf("unsupported distribution channel %q", in.Channel)
	}
	return plan, nil
}
