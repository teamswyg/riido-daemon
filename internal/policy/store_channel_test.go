package policy_test

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/policy"
)

func TestStoreChannelPolicyAlwaysDeniesForbiddenSurfaces(t *testing.T) {
	channels := []hostintegration.DistributionChannel{
		hostintegration.DistributionChannelDeveloperID,
		hostintegration.DistributionChannelMacAppStore,
		hostintegration.DistributionChannelMSIXSideload,
		hostintegration.DistributionChannelMSIXStore,
	}
	surfaces := []policy.StoreSurface{
		policy.StoreSurfaceProviderCLIBundling,
		policy.StoreSurfaceSilentProviderAutoInstall,
		policy.StoreSurfaceExternalTCPListener,
		policy.StoreSurfaceArbitraryHomeScan,
	}

	for _, channel := range channels {
		for _, surface := range surfaces {
			decision := policy.EvaluateStoreChannelPolicy(policy.StoreChannelPolicyInput{
				Channel: channel,
				Surface: surface,
			})
			if decision.Allowed {
				t.Fatalf("%s/%s allowed, want denied", channel, surface)
			}
			if decision.Code != "STORE_CHANNEL_SURFACE_FORBIDDEN" {
				t.Fatalf("%s/%s code = %q", channel, surface, decision.Code)
			}
		}
	}
}

func TestStoreChannelPolicyMacAppStoreProviderPathRequiresOSGrantAndStoreReview(t *testing.T) {
	withoutGrant := policy.EvaluateStoreChannelPolicy(policy.StoreChannelPolicyInput{
		Channel: hostintegration.DistributionChannelMacAppStore,
		Surface: policy.StoreSurfaceProviderCLIUserSelectedPath,
	})
	if withoutGrant.Allowed {
		t.Fatal("mac-app-store provider path without OS grant allowed, want denied")
	}
	if withoutGrant.Code != "STORE_CHANNEL_REQUIRES_OS_GRANT" {
		t.Fatalf("code = %q", withoutGrant.Code)
	}

	withGrant := policy.EvaluateStoreChannelPolicy(policy.StoreChannelPolicyInput{
		Channel:        hostintegration.DistributionChannelMacAppStore,
		Surface:        policy.StoreSurfaceProviderCLIUserSelectedPath,
		OSGrantPresent: true,
	})
	if withGrant.Allowed || withGrant.Code != "STORE_CHANNEL_REQUIRES_STORE_REVIEW_APPROVAL" {
		t.Fatalf("mac-app-store provider path without review approval decision = %+v", withGrant)
	}

	withGrantAndReview := policy.EvaluateStoreChannelPolicy(policy.StoreChannelPolicyInput{
		Channel:             hostintegration.DistributionChannelMacAppStore,
		Surface:             policy.StoreSurfaceProviderCLIUserSelectedPath,
		OSGrantPresent:      true,
		StoreReviewApproved: true,
	})
	if !withGrantAndReview.Allowed {
		t.Fatalf("mac-app-store provider path with OS grant/review denied: %s", withGrantAndReview.Reason)
	}
}

func TestStoreChannelPolicyBackgroundHelperRequiresConsentAndStoreReview(t *testing.T) {
	withoutConsent := policy.EvaluateStoreChannelPolicy(policy.StoreChannelPolicyInput{
		Channel: hostintegration.DistributionChannelDeveloperID,
		Surface: policy.StoreSurfaceBackgroundHelper,
	})
	if withoutConsent.Allowed || withoutConsent.Code != "STORE_CHANNEL_REQUIRES_CONSENT" {
		t.Fatalf("developer-id background helper decision = %+v", withoutConsent)
	}

	withoutReview := policy.EvaluateStoreChannelPolicy(policy.StoreChannelPolicyInput{
		Channel:                hostintegration.DistributionChannelMSIXStore,
		Surface:                policy.StoreSurfaceBackgroundHelper,
		ExplicitConsentGranted: true,
	})
	if withoutReview.Allowed || withoutReview.Code != "STORE_CHANNEL_REQUIRES_STORE_REVIEW_APPROVAL" {
		t.Fatalf("msix-store background helper decision = %+v", withoutReview)
	}

	withReview := policy.EvaluateStoreChannelPolicy(policy.StoreChannelPolicyInput{
		Channel:                hostintegration.DistributionChannelMSIXStore,
		Surface:                policy.StoreSurfaceBackgroundHelper,
		ExplicitConsentGranted: true,
		StoreReviewApproved:    true,
	})
	if !withReview.Allowed {
		t.Fatalf("msix-store background helper with consent/review denied: %s", withReview.Reason)
	}
}

func TestStoreChannelPolicyLocalIPCAllowedAcrossStoreChannels(t *testing.T) {
	channels := []hostintegration.DistributionChannel{
		hostintegration.DistributionChannelDeveloperID,
		hostintegration.DistributionChannelMacAppStore,
		hostintegration.DistributionChannelMSIXSideload,
		hostintegration.DistributionChannelMSIXStore,
	}

	for _, channel := range channels {
		decision := policy.EvaluateStoreChannelPolicy(policy.StoreChannelPolicyInput{
			Channel: channel,
			Surface: policy.StoreSurfaceLocalIPC,
		})
		if !decision.Allowed {
			t.Fatalf("%s local IPC denied: %s", channel, decision.Reason)
		}
	}
}

func TestStoreChannelPolicyStoreManagedChannelsDenySelfUpdater(t *testing.T) {
	for _, channel := range []hostintegration.DistributionChannel{
		hostintegration.DistributionChannelMacAppStore,
		hostintegration.DistributionChannelMSIXStore,
	} {
		decision := policy.EvaluateStoreChannelPolicy(policy.StoreChannelPolicyInput{
			Channel: channel,
			Surface: policy.StoreSurfaceSelfUpdater,
		})
		if decision.Allowed {
			t.Fatalf("%s self-updater allowed, want store update mechanism only", channel)
		}
	}
}
