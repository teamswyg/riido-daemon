package runtimeactor

import (
	"context"
	"path/filepath"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/detectutil"
)

// This interface is implemented only by providers that support per-run executable selection.
type resumeExecutableAdapter interface {
	ResumeExecutableIdentity(path, version string) agentbridge.DetectResult
	ProbeResumeExecutable(context.Context, string) (agentbridge.DetectResult, error)
}

type NativeResumeRuntime struct {
	owner      *Actor
	capability Capability
	probe      resumeExecutableAdapter
}

func (r *NativeResumeRuntime) Capability() Capability { return r.capability }

// Match the retained fingerprint offline before invoking any candidate, then
// verify its actual reported version, authentication and protocol capabilities.
func (a *Actor) ResolveNativeResume(ctx context.Context, provider, fingerprint, path, version string) (*NativeResumeRuntime, error) {
	adapter, ok := indexAdapters(a.cfg.Adapters)[provider].(resumeExecutableAdapter)
	if !ok || version == "" || fingerprint == "" {
		return nil, ErrRuntimePinViolated
	}
	candidates := append([]string{path}, detectutil.ResolveExecutableCandidates(provider, "")...)
	seen := map[string]bool{}
	for _, candidate := range candidates {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if !filepath.IsAbs(candidate) || seen[candidate] {
			continue
		}
		seen[candidate] = true
		identity := adapter.ResumeExecutableIdentity(candidate, version)
		capability, err := buildRuntimeCapability(a.cfg.RuntimeID, provider, identity, a.cfg.PolicyBundleVersion, a.cfg.Now())
		if err != nil || capability.DetectedFingerprint == "" || capability.CapabilityFingerprint != fingerprint {
			continue
		}
		runtime := &NativeResumeRuntime{owner: a, capability: capability, probe: adapter}
		observed, err := runtime.observe(ctx, a)
		if err == nil && observed.Available && observed.CapabilityFingerprint == fingerprint {
			runtime.capability = observed
			return runtime, nil
		}
	}
	return nil, ErrRuntimePinViolated
}

func (r *NativeResumeRuntime) observe(ctx context.Context, a *Actor) (Capability, error) {
	c := r.capability
	if string(detectedFingerprintForExecutable(c.Executable)) != c.DetectedFingerprint {
		return Capability{}, ErrRuntimePinViolated
	}
	res, err := r.probe.ProbeResumeExecutable(ctx, c.Executable)
	if err != nil {
		return Capability{}, err
	}
	return buildRuntimeCapability(a.cfg.RuntimeID, c.Provider, res, a.cfg.PolicyBundleVersion, a.cfg.Now())
}

// The actor's normal probe observes the default CLI. An alternate installation
// must fence its own running task if its binary or authenticated identity drifts.
func (r *NativeResumeRuntime) watch(ctx context.Context, a *Actor, h *SessionHandle) {
	if ctx == nil {
		ctx = context.Background()
	}
	if a.cfg.CapabilityRefreshEvery <= 0 {
		return
	}
	timer := time.NewTicker(a.cfg.CapabilityRefreshEvery)
	defer timer.Stop()
	for {
		select {
		case <-h.Done():
			return
		case <-ctx.Done():
			return
		case <-timer.C:
			probeCtx, cancel := context.WithTimeout(ctx, DefaultCapabilityProbeTimeout)
			observed, err := r.observe(probeCtx, a)
			cancel()
			if err != nil || observed.CapabilityFingerprint != r.capability.CapabilityFingerprint || !observed.Available {
				h.session.TerminateWithContext(ctx, agentbridge.Result{Status: agentbridge.ResultBlocked, Error: ErrRuntimePinViolated.Error()})
				return
			}
		}
	}
}
