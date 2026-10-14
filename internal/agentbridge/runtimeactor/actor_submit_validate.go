package runtimeactor

import (
	"errors"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (a *Actor) validateSubmit(inFlight map[string]*runningTask, msg *submitMsg) error {
	if msg.req.ID == "" {
		return errors.New("runtimeactor: TaskRequest.ID is required")
	}
	if _, dup := inFlight[msg.req.ID]; dup {
		return fmt.Errorf("%w: %s", ErrDuplicateTaskID, msg.req.ID)
	}
	if len(inFlight) >= a.cfg.MaxConcurrent {
		return ErrSlotExhausted
	}
	return nil
}

func (a *Actor) submitCapability(
	msg *submitMsg,
	adapters map[string]agentbridge.Adapter,
	caps []Capability,
	detectedAt map[string]time.Time,
) (agentbridge.Adapter, Capability, error) {
	adapter, ok := adapters[string(msg.req.Provider)]
	if !ok {
		return nil, Capability{}, fmt.Errorf("%w: %s", ErrUnknownProvider, msg.req.Provider)
	}
	capView, ok, err := a.capabilityForSubmit(msg.ctx, adapter, caps, detectedAt, string(msg.req.Provider))
	if err != nil {
		return nil, Capability{}, err
	}
	if !ok || !capView.Available {
		return nil, Capability{}, fmt.Errorf("%w: %s", ErrProviderUnavailable, msg.req.Provider)
	}
	return adapter, capView, nil
}
