package scheduling

func supportsSurface(candidate RuntimeCapability, surface RequiredSurface) (bool, bool) {
	switch surface {
	case SurfaceStructuredEventStream:
		return candidate.SupportsStreaming, true
	case SurfaceSessionResume:
		return candidate.SupportsResume, true
	case SurfaceSystemPrompt:
		return candidate.SupportsSystem, true
	case SurfaceMaxTurns:
		return candidate.SupportsMaxTurns, true
	case SurfaceMCP:
		return candidate.SupportsMCP, true
	case SurfaceToolHooks:
		return candidate.SupportsToolHooks, true
	case SurfaceUsage:
		return candidate.SupportsUsage, true
	case SurfaceWorktree:
		return candidate.SupportsWorktree, true
	default:
		return false, false
	}
}
