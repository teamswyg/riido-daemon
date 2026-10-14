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
