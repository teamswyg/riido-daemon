package main

import "os"

func observeProvider(root string, provider provider, runIntegration bool) providerEvidence {
	override := os.Getenv(provider.OverrideEnv)
	exe, found := resolveExecutable(provider.DefaultExecutable, override)
	ev := providerEvidence{
		ID:                 provider.ID,
		Available:          found,
		ExecutableRef:      executableRef(provider, override, found),
		IntegrationCommand: integrationCommand(provider),
	}
	if !found {
		ev.IntegrationStatus = "skipped"
		ev.FailureSummary = "executable not found on PATH or override env"
		ev.Repair = repairPointer(classifyRepair(provider.ID, ev.IntegrationStatus, ev.FailureSummary, ev.Available))
		return ev
	}
	ev.Version = probeVersion(exe)
	if !runIntegration {
		ev.IntegrationStatus = "observed"
		return ev
	}
	status, failure := runIntegrationTest(root, provider)
	ev.IntegrationStatus = status
	ev.FailureSummary = failure
	ev.Repair = repairPointer(classifyRepair(provider.ID, status, failure, ev.Available))
	return ev
}

func repairPointer(rep repair) *repair {
	if rep.Class == "" {
		return nil
	}
	return &rep
}
