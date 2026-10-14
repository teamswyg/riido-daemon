package hostintegration

import "testing"

func TestDefaultLocalIPCEndpointWindowsUsesNamedPipe(t *testing.T) {
	root, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:                     DistributionChannelMSIXStore,
		HostOS:                      HostOSWindows,
		WindowsPackageLocalDataRoot: `C:\Users\tester\AppData\Local\Packages\Riido_abc\LocalState`,
	})
	if err != nil {
		t.Fatal(err)
	}

	endpoint, err := DefaultLocalIPCEndpoint(LocalIPCEndpointInput{
		Channel:     DistributionChannelMSIXStore,
		HostOS:      HostOSWindows,
		AppDataRoot: root,
		Owner:       LocalIPCOwnerHelper,
		Name:        "riido",
	})
	if err != nil {
		t.Fatal(err)
	}

	if endpoint.EndpointKind != LocalIPCEndpointNamedPipe {
		t.Fatalf("kind = %q, want %q", endpoint.EndpointKind, LocalIPCEndpointNamedPipe)
	}
	if endpoint.Path != `\\.\pipe\riido-msix-store-helper-riido` {
		t.Fatalf("path = %q", endpoint.Path)
	}
}
