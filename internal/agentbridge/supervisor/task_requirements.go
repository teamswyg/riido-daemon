package supervisor

import (
	"strings"

	"github.com/teamswyg/riido-contracts/provider/capability"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
	"github.com/teamswyg/riido-daemon/internal/scheduling"
)

func taskRequirements(req *bridge.TaskRequest) scheduling.TaskRequirements {
	surfaces := make([]scheduling.RequiredSurface, 0, len(req.RequiredSurfaces))
	for _, surface := range req.RequiredSurfaces {
		surfaces = append(surfaces, scheduling.RequiredSurface(surface))
	}
	surfaces = append(surfaces, metadataRequiredSurfaces(req.Metadata)...)
	if req.Worktree != nil {
		surfaces = append(surfaces, scheduling.SurfaceWorktree)
	}
	return scheduling.TaskRequirements{
		Provider:                 capability.ProviderKind(req.Provider),
		RequiredSurfaces:         scheduling.NormalizeRequiredSurfaces(surfaces),
		AllowExperimentalRuntime: req.AllowExperimentalRuntime || metadataBool(req.Metadata, MetadataAllowExperimentalRuntime),
	}
}

func metadataRequiredSurfaces(meta map[string]string) []scheduling.RequiredSurface {
	if meta == nil {
		return nil
	}
	var surfaces []scheduling.RequiredSurface
	for surface := range strings.SplitSeq(meta[MetadataRequiredSurfaces], ",") {
		surfaces = append(surfaces, scheduling.RequiredSurface(surface))
	}
	return surfaces
}

func metadataBool(meta map[string]string, key string) bool {
	if meta == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(meta[key])) {
	case "1", "true", "yes", "y":
		return true
	default:
		return false
	}
}
