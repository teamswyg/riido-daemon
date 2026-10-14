package supervisor

import (
	"path/filepath"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/workdir"
)

func applyPreparedWorkspaceMetadata(req *bridge.TaskRequest, ws workdir.Workspace, runID string, nativePlan workdir.ProviderNativeConfigPlan, nativeConfigVersion string) {
	req.Cwd = ws.Workdir
	req.Metadata[MetadataRunID] = runID
	req.Metadata[MetadataWorkdirRoot] = ws.Root
	req.Metadata[MetadataWorkdir] = ws.Workdir
	req.Metadata[MetadataOutputDir] = ws.Output
	req.Metadata[MetadataLogsDir] = ws.Logs
	req.Metadata[MetadataArtifactsDir] = ws.Artifacts
	req.Metadata[MetadataNativeConfig] = ws.NativeConfig
	if nativePlan.ConfigHomeDir != "" {
		req.Metadata[MetadataNativeConfigHome] = filepath.Join(ws.Workdir, filepath.FromSlash(nativePlan.ConfigHomeDir))
	} else {
		delete(req.Metadata, MetadataNativeConfigHome)
	}
	req.Metadata[MetadataIRDir] = ws.IR
	req.Metadata[MetadataNativeConfigVersion] = nativeConfigVersion
}
