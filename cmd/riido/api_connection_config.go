package main

import (
	"fmt"
	"os"

	"github.com/teamswyg/riido-daemon/internal/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/riidoapi"
)

type apiCLIConfig struct {
	socketPath string
	transport  riidoapi.LocalTransport
}

func defaultAPICLIConfig() (apiCLIConfig, error) {
	return defaultAPICLIConfigForOS(currentDaemonHostOS(), os.UserHomeDir, os.UserCacheDir)
}

func defaultAPICLIConfigForOS(
	goos string,
	userHome daemonPathRootResolver,
	userCache daemonPathRootResolver,
) (apiCLIConfig, error) {
	input, err := defaultDaemonAppDataRootInput(goos, userHome, userCache)
	if err != nil {
		return apiCLIConfig{}, err
	}
	root, err := hostintegration.DefaultAppDataRoot(input)
	if err != nil {
		return apiCLIConfig{}, fmt.Errorf("resolve default API app data root: %w", err)
	}
	endpoint, err := hostintegration.DefaultLocalIPCEndpoint(hostintegration.LocalIPCEndpointInput{
		Channel:     hostintegration.DistributionChannelDevLocal,
		HostOS:      root.HostOS,
		AppDataRoot: root,
		Owner:       hostintegration.LocalIPCOwnerHelper,
		Name:        "riido",
	})
	if err != nil {
		return apiCLIConfig{}, fmt.Errorf("resolve default API local endpoint: %w", err)
	}
	transport, err := apiTransportForEndpointKind(endpoint.EndpointKind)
	if err != nil {
		return apiCLIConfig{}, err
	}
	return apiCLIConfig{socketPath: endpoint.Path, transport: transport}, nil
}

func apiTransportForEndpointKind(kind hostintegration.LocalIPCEndpointKind) (riidoapi.LocalTransport, error) {
	switch kind {
	case hostintegration.LocalIPCEndpointUnixSocket:
		return riidoapi.LocalTransportUnixSocket, nil
	case hostintegration.LocalIPCEndpointNamedPipe:
		return riidoapi.LocalTransportWindowsNamedPipe, nil
	default:
		return "", fmt.Errorf("unsupported API local endpoint kind %q", kind)
	}
}
