package supervisor

import "github.com/teamswyg/riido-contracts/metadatakeys"

const (
	MetadataWorkspaceID              = string(metadatakeys.WorkspaceID)
	MetadataWorkspace                = string(metadatakeys.Workspace)
	MetadataRunID                    = string(metadatakeys.RunID)
	MetadataAgentName                = string(metadatakeys.AgentName)
	MetadataAgentIdentity            = string(metadatakeys.AgentIdentity)
	MetadataWorkflow                 = string(metadatakeys.Workflow)
	MetadataWorkdirRoot              = string(metadatakeys.WorkdirRoot)
	MetadataWorkdir                  = string(metadatakeys.Workdir)
	MetadataOutputDir                = string(metadatakeys.OutputDir)
	MetadataLogsDir                  = string(metadatakeys.LogsDir)
	MetadataArtifactsDir             = string(metadatakeys.ArtifactsDir)
	MetadataNativeConfig             = string(metadatakeys.NativeConfigDir)
	MetadataNativeConfigHome         = string(metadatakeys.NativeConfigHome)
	MetadataIRDir                    = string(metadatakeys.IRDir)
	MetadataNativeConfigVersion      = string(metadatakeys.NativeConfigVersion)
	MetadataRequiredSurfaces         = string(metadatakeys.RequiredSurfaces)
	MetadataAllowExperimentalRuntime = string(metadatakeys.AllowExperimentalRuntime)
)
