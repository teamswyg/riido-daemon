package main

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/riidoapi"
)

func requestAPI(
	config apiCLIConfig,
	timeout time.Duration,
	method riidoapi.Method,
	request any,
	response any,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client := riidoapi.NewClientWithTransport(config.transport, config.socketPath)
	client.Timeout = timeout
	return client.Request(ctx, string(method), request, response)
}
