package hostintegration

import "testing"

func TestDefaultAppDataRootMSIXStoreRequiresPackageLocalRoot(t *testing.T) {
	_, err := DefaultAppDataRoot(AppDataRootInput{
		Channel: DistributionChannelMSIXStore,
		HostOS:  HostOSWindows,
	})
	if err == nil {
		t.Fatal("expected msix store root without package local data to fail")
	}
}

func TestDefaultAppDataRootMSIXStoreUsesWindowsSeparators(t *testing.T) {
	root, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:                     DistributionChannelMSIXStore,
		HostOS:                      HostOSWindows,
		WindowsPackageLocalDataRoot: `C:\Users\tester\AppData\Local\Packages\Riido_abc\LocalState`,
	})
	if err != nil {
		t.Fatal(err)
	}

	if root.Scope != AppDataRootWindowsPackageLocal {
		t.Fatalf("scope = %q, want %q", root.Scope, AppDataRootWindowsPackageLocal)
	}
	if root.WorkdirRoot() != `C:\Users\tester\AppData\Local\Packages\Riido_abc\LocalState\workspaces` {
		t.Fatalf("workdir root = %q", root.WorkdirRoot())
	}
}
