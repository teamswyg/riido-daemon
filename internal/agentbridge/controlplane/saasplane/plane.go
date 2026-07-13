package saasplane

import (
	"net/http"
	"sync"
)

// Plane implements TaskSourcePort and TaskReporterPort against SaaS assignment
// polling. Internal state is owned by a mailbox goroutine.
type Plane struct {
	cfg                      Config
	client                   *http.Client
	ops                      chan stateOp
	done                     chan struct{}
	deviceCredentialRejected chan error
	deviceCredentialOnce     sync.Once
}

func (p *Plane) Close() {
	ack := make(chan struct{})
	select {
	case p.ops <- stateOp{close: true, ack: ack}:
		<-ack
	case <-p.done:
	}
}

// DeviceCredentialRejected reports the first HTTP 401 received while the
// plane is using a device credential. Managed launchers use this signal to
// replace the process and reload a credential rotated by Desktop.
func (p *Plane) DeviceCredentialRejected() <-chan error {
	if p == nil {
		return nil
	}
	return p.deviceCredentialRejected
}
