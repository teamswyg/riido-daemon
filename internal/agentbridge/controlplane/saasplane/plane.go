package saasplane

import (
	"net/http"
	"sync"
)

// Plane implements TaskSourcePort and TaskReporterPort against SaaS assignment
// polling. Internal state is owned by a mailbox goroutine.
type Plane struct {
	notices    assignmentNotifications
	noticeOnce sync.Once
	cfg        Config
	client     *http.Client
	ops        chan stateOp
	done       chan struct{}
}

func (p *Plane) Close() {
	p.stopAssignmentNotifications()
	ack := make(chan struct{})
	select {
	case p.ops <- stateOp{close: true, ack: ack}:
		<-ack
	case <-p.done:
	}
}
