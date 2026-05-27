package hostintegration

import "testing"

func TestDefaultAppDataRootDarwinDevLocal(t *testing.T) {
	root, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:  DistributionChannelDevLocal,
		HostOS:   HostOSDarwin,
		UserHome: "/Users/tester",
	})
	if err != nil {
		t.Fatal(err)
	}

	if root.Scope != AppDataRootUserApplicationSupport {
		t.Fatalf("scope = %q, want %q", root.Scope, AppDataRootUserApplicationSupport)
	}
	if root.Path != "/Users/tester/Library/Application Support/riido" {
		t.Fatalf("path = %q", root.Path)
	}
	if root.WorkdirRoot() != "/Users/tester/Library/Application Support/riido/workspaces" {
		t.Fatalf("workdir root = %q", root.WorkdirRoot())
	}
}

func TestDefaultAppDataRootMacAppStoreRequiresStoreRoot(t *testing.T) {
	_, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:  DistributionChannelMacAppStore,
		HostOS:   HostOSDarwin,
		UserHome: "/Users/tester",
	})
	if err == nil {
		t.Fatal("expected mac app store root without app group/container to fail")
	}
}

func TestDefaultAppDataRootMacAppStorePrefersAppGroup(t *testing.T) {
	root, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:                    DistributionChannelMacAppStore,
		HostOS:                     HostOSDarwin,
		DarwinSandboxContainerRoot: "/Users/tester/Library/Containers/io.riido.app/Data",
		DarwinAppGroupRoot:         "/Users/tester/Library/Group Containers/group.io.riido",
	})
	if err != nil {
		t.Fatal(err)
	}

	if root.Scope != AppDataRootAppGroup {
		t.Fatalf("scope = %q, want %q", root.Scope, AppDataRootAppGroup)
	}
	if root.Path != "/Users/tester/Library/Group Containers/group.io.riido" {
		t.Fatalf("path = %q", root.Path)
	}
}

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

func TestDefaultAppDataRootRejectsChannelOSMismatch(t *testing.T) {
	_, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:                     DistributionChannelMacAppStore,
		HostOS:                      HostOSWindows,
		WindowsPackageLocalDataRoot: `C:\Users\tester\AppData\Local\Packages\Riido_abc\LocalState`,
	})
	if err == nil {
		t.Fatal("expected mac app store on windows to fail")
	}
}
