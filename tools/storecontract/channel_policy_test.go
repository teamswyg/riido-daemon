package main

import (
	"testing"
)

func TestRunRejectsMacAppStorePolicyDrift(t *testing.T) {
	cases := []contractMutationCase{
		{
			name:   "missing self updater ban",
			mutate: removeChannelForbiddenSurface("mac-app-store", "self-updater"),
			error:  `channel "mac-app-store" must forbid self-updater`,
		},
		{
			name:   "wrong runtime role",
			mutate: setChannelRuntimeRole("mac-app-store", "local-helper-broker"),
			error:  `channel "mac-app-store" runtime_role must be sandboxed-login-item-helper`,
		},
		{
			name:   "wrong background rule",
			mutate: setChannelBackgroundRule("mac-app-store", "explicit-consent"),
			error:  `channel "mac-app-store" background_rule must be explicit-consent-and-store-review`,
		},
		{
			name:   "wrong update mechanism",
			mutate: setChannelUpdateMechanism("mac-app-store", "self-managed"),
			error:  `channel "mac-app-store" update_mechanism must be app-store-managed`,
		},
	}
	expectContractMutationFailures(t, cases)
}

func TestRunRejectsMSIXStorePolicyDrift(t *testing.T) {
	cases := []contractMutationCase{
		{
			name:   "missing review note",
			mutate: removeChannelRequiredSurface("msix-store", "runfulltrust-review-note"),
			error:  `channel "msix-store" must require runfulltrust-review-note`,
		},
		{
			name:   "wrong runtime role",
			mutate: setChannelRuntimeRole("msix-store", "msix-packaged-helper-broker"),
			error:  `channel "msix-store" runtime_role must be msix-packaged-full-trust-helper-tray`,
		},
		{
			name:   "wrong background rule",
			mutate: setChannelBackgroundRule("msix-store", "explicit-consent"),
			error:  `channel "msix-store" background_rule must be explicit-consent-and-store-review`,
		},
		{
			name:   "wrong update mechanism",
			mutate: setChannelUpdateMechanism("msix-store", "self-managed"),
			error:  `channel "msix-store" update_mechanism must be store-managed`,
		},
		{
			name:   "policy gate not closed",
			mutate: setChannelStatus("msix-store", "requires-policy-gate"),
			error:  `channel "msix-store" status must be store-review-ready`,
		},
	}
	expectContractMutationFailures(t, cases)
}
