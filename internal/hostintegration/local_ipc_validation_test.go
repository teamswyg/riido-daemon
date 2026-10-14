package hostintegration

import "testing"

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
