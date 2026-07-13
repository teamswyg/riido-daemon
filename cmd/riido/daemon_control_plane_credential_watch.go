package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/logging"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

type daemonDeviceCredentialRejectionSource interface {
	DeviceCredentialRejected() <-chan error
}

func watchDaemonDeviceCredentialRejection(
	ctx lifecycle.Context,
	source any,
	cancelDaemon context.CancelFunc,
	log logging.Logger,
) context.CancelFunc {
	watchCtx, stop := context.WithCancel(ctx.Context())
	rejections, ok := source.(daemonDeviceCredentialRejectionSource)
	if !ok || rejections.DeviceCredentialRejected() == nil {
		return stop
	}
	go func() {
		select {
		case err := <-rejections.DeviceCredentialRejected():
			if err == nil {
				return
			}
			log.Printf("device credential rejected; stopping daemon to reload managed credential: %v", err)
			cancelDaemon()
		case <-watchCtx.Done():
		}
	}()
	return stop
}
