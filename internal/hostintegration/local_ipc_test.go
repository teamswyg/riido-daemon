package hostintegration

import "testing"

func TestDefaultLocalIPCEndpointDarwinUsesAppDataUnixSocket(t *testing.T) {
	root, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:  DistributionChannelDevLocal,
		HostOS:   HostOSDarwin,
		UserHome: "/Users/tester",
	})
	if err != nil {
		t.Fatal(err)
	}

	endpoint, err := DefaultLocalIPCEndpoint(LocalIPCEndpointInput{
		Channel:     DistributionChannelDevLocal,
		HostOS:      HostOSDarwin,
		AppDataRoot: root,
		Owner:       LocalIPCOwnerHelper,
		Name:        "agentd",
	})
	if err != nil {
		t.Fatal(err)
	}

	if endpoint.EndpointKind != LocalIPCEndpointUnixSocket {
		t.Fatalf("kind = %q, want %q", endpoint.EndpointKind, LocalIPCEndpointUnixSocket)
	}
	if endpoint.Path != "/Users/tester/Library/Application Support/riido/agentd.sock" {
		t.Fatalf("path = %q", endpoint.Path)
	}
}

func TestDefaultLocalIPCEndpointMacAppStoreUsesContainerUnixSocket(t *testing.T) {
	root, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:                    DistributionChannelMacAppStore,
		HostOS:                     HostOSDarwin,
		DarwinSandboxContainerRoot: "/Users/tester/Library/Containers/io.riido.app/Data",
	})
	if err != nil {
		t.Fatal(err)
	}

	endpoint, err := DefaultLocalIPCEndpoint(LocalIPCEndpointInput{
		Channel:     DistributionChannelMacAppStore,
		HostOS:      HostOSDarwin,
		AppDataRoot: root,
		Owner:       LocalIPCOwnerHelper,
	})
	if err != nil {
		t.Fatal(err)
	}

	if endpoint.Path != "/Users/tester/Library/Containers/io.riido.app/Data/riido.sock" {
		t.Fatalf("path = %q", endpoint.Path)
	}
}

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

func TestDefaultLocalIPCEndpointRejectsRootMismatch(t *testing.T) {
	root, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:  DistributionChannelDevLocal,
		HostOS:   HostOSDarwin,
		UserHome: "/Users/tester",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = DefaultLocalIPCEndpoint(LocalIPCEndpointInput{
		Channel:     DistributionChannelMacAppStore,
		HostOS:      HostOSDarwin,
		AppDataRoot: root,
		Owner:       LocalIPCOwnerHelper,
	})
	if err == nil {
		t.Fatal("expected root/channel mismatch to fail")
	}
}

func TestDefaultLocalIPCEndpointRejectsUnsafeName(t *testing.T) {
	root, err := DefaultAppDataRoot(AppDataRootInput{
		Channel:  DistributionChannelDevLocal,
		HostOS:   HostOSDarwin,
		UserHome: "/Users/tester",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = DefaultLocalIPCEndpoint(LocalIPCEndpointInput{
		Channel:     DistributionChannelDevLocal,
		HostOS:      HostOSDarwin,
		AppDataRoot: root,
		Owner:       LocalIPCOwnerHelper,
		Name:        "../riido",
	})
	if err == nil {
		t.Fatal("expected unsafe endpoint name to fail")
	}
}
