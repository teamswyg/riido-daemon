package main

import "github.com/teamswyg/riido-daemon/internal/riidoapi"

type apiCLIConfig struct {
	socketPath string
	transport  riidoapi.LocalTransport
}

func defaultAPICLIConfig() (apiCLIConfig, error) {
	socketPath, err := riidoapi.DefaultSocketPath()
	if err != nil {
		return apiCLIConfig{}, err
	}
	return apiCLIConfig{socketPath: socketPath, transport: riidoapi.LocalTransportUnixSocket}, nil
}
