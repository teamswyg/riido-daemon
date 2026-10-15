package supervisor

import (
	"context"
	"errors"
	"maps"

	"github.com/teamswyg/riido-contracts/assignment"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func (a *Actor) recoveryFallback(ctx context.Context, original runtimeactor.Status, req *bridge.TaskRequest, candidates []assignment.ProviderRecoveryCandidate) (*runtimeactor.Actor, runtimeactor.Status, *bridge.TaskRequest, error) {
	if req.ResumeSessionID != "" || req.ProviderNeutralPrompt == "" || len(req.Env) > 0 || len(req.CustomArgs) > 0 || len(req.MCPConfig) > 0 {
		return nil, runtimeactor.Status{}, nil, errors.New("handoff has native session or provider-specific authority")
	}
	for _, candidate := range candidates {
		if candidate.Provider == string(req.Provider) {
			continue
		}
		for _, rt := range configuredRuntimes(a.cfg) {
			status, err := rt.Status(ctx)
			if err != nil || status.Owner != original.Owner || status.DeviceName != original.DeviceName {
				continue
			}
			capView, ok := findCapability(status.Capabilities, candidate.Provider)
			if !ok || !capView.Available || capView.HealthStatus != "healthy" {
				continue
			}
			modelFound := false
			for _, model := range status.Models {
				if model.ModelID == candidate.ModelID {
					modelFound = true
				}
			}
			if !modelFound {
				continue
			}
			next := *req
			next.Metadata = maps.Clone(req.Metadata)
			next.Provider = bridge.Provider(candidate.Provider)
			next.ExpectedCapabilityFingerprint = capView.CapabilityFingerprint
			next.RequireHealthyCapability = true
			next.Model = providercatalog.ModelOverride(candidate.Provider, candidate.ModelID)
			prompt, system, placement, instruction := agentbridge.ApplyRuntimeInstructionContract(candidate.Provider, req.ProviderNeutralPrompt, "", req.AgentInstruction)
			next.Prompt = prompt
			next.SystemPrompt = system
			next.Metadata[agentbridge.MetadataTelemetryContract] = placement
			if instruction != "" {
				next.Metadata[agentbridge.MetadataAgentInstruction] = instruction
			} else {
				delete(next.Metadata, agentbridge.MetadataAgentInstruction)
			}
			next.CustomArgs = nil
			next.MCPConfig = nil
			next.Env = nil
			next.ResumeSessionID = ""
			if !taskEligibility(status, &next).Eligible {
				continue
			}
			return rt, status, &next, nil
		}
	}
	return nil, runtimeactor.Status{}, nil, errors.New("no authorized compatible handoff target")
}

func (a *Actor) reconfigureRecoveryWorkspace(status runtimeactor.Status, req *bridge.TaskRequest, prepared *preparedWorkspace) error {
	if prepared == nil {
		return nil
	}
	if prepared.workspace == nil {
		return errors.New("handoff workspace is not retained")
	}
	capView, err := workspacePrepareCapability(status, req)
	if err != nil {
		return err
	}
	native, err := a.resolveNativeWorkspaceConfig(req)
	if err != nil {
		return err
	}
	if err = a.injectWorkspaceRuntimeConfig(*prepared.workspace, req, capView.ProtocolKind, native); err != nil {
		return err
	}
	version, err := workdir.ComputeNativeConfigVersion(*prepared.workspace, workdir.NativeConfigVersionInput{PolicyBundleVersion: a.cfg.PolicyBundleVersion, ProviderKind: string(req.Provider), ProtocolKind: capView.ProtocolKind})
	if err != nil {
		return err
	}
	cwd := req.Cwd
	applyPreparedWorkspaceMetadata(req, *prepared.workspace, req.Metadata[MetadataRunID], native.resolved, version)
	req.Cwd = cwd
	if prepared.execution != nil {
		appendRepositoryInstructions(req, prepared.execution.Binding)
	}
	if prepared.events != nil {
		prepared.events.runtimeID = status.RuntimeID
		prepared.events.capability = capView
		prepared.events.nativeConfigVersion = version
	}
	prepared.canResume = capView.SupportsResume
	return nil
}
