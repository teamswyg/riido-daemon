package hostintegration

import (
	"errors"
	"fmt"
)

// DefaultAppDataRoot resolves the default app data root for a distribution
// channel. Store-managed channels require adapter-provided container/package
// roots and never derive from arbitrary user home scanning.
func DefaultAppDataRoot(in AppDataRootInput) (AppDataRoot, error) {
	if !in.Channel.Valid() {
		return AppDataRoot{}, fmt.Errorf("unknown distribution channel %q", in.Channel)
	}
	if !in.HostOS.Valid() {
		return AppDataRoot{}, fmt.Errorf("unknown host OS %q", in.HostOS)
	}

	switch in.Channel {
	case DistributionChannelDevLocal:
		return devLocalAppDataRoot(in)
	case DistributionChannelDeveloperID:
		if in.HostOS != HostOSDarwin {
			return AppDataRoot{}, errors.New("developer-id app data root requires darwin host OS")
		}
		return darwinApplicationSupportRoot(in)
	case DistributionChannelMacAppStore:
		if in.HostOS != HostOSDarwin {
			return AppDataRoot{}, errors.New("mac-app-store app data root requires darwin host OS")
		}
		return darwinStoreAppDataRoot(in)
	case DistributionChannelMSIXSideload, DistributionChannelMSIXStore:
		if in.HostOS != HostOSWindows {
			return AppDataRoot{}, errors.New("msix app data root requires windows host OS")
		}
		return windowsPackageLocalRoot(in)
	default:
		return AppDataRoot{}, fmt.Errorf("unsupported distribution channel %q", in.Channel)
	}
}
