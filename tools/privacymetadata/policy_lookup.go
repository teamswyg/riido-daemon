package main

func findSurface(policy PolicySnapshot, id string) (SurfaceSnapshot, bool) {
	for _, surface := range policy.Surfaces {
		if surface.ID == id {
			return surface, true
		}
	}
	return SurfaceSnapshot{}, false
}
