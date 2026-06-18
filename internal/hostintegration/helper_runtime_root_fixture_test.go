package hostintegration

import "testing"

func mustMSIXAppDataRoot(t *testing.T, channel DistributionChannel) AppDataRoot {
	t.Helper()
	root, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:                     channel,
		HostOS:                      HostOSWindows,
		WindowsPackageLocalDataRoot: `C:\Users\tester\AppData\Local\Packages\Riido_abc\LocalState`,
	})
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func mustDarwinStoreAppDataRoot(t *testing.T, channel DistributionChannel) AppDataRoot {
	t.Helper()
	root, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:            channel,
		HostOS:             HostOSDarwin,
		DarwinAppGroupRoot: "/Users/tester/Library/Group Containers/group.io.riido",
	})
	if err != nil {
		t.Fatal(err)
	}
	return root
}
