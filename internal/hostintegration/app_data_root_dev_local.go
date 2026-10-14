package hostintegration

import "fmt"

func devLocalAppDataRoot(in AppDataRootInput) (AppDataRoot, error) {
	switch in.HostOS {
	case HostOSDarwin:
		return darwinApplicationSupportRoot(in)
	case HostOSWindows:
		return windowsLocalAppDataRoot(in)
	default:
		return AppDataRoot{}, fmt.Errorf("unsupported dev-local host OS %q", in.HostOS)
	}
}
