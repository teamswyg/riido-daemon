package main

func removeChannelRequiredSurface(id, surface string) func(*contract) {
	return mutateChannel(id, func(item *channel) {
		item.RequiredSurfaces = removeString(item.RequiredSurfaces, surface)
	})
}

func removeChannelForbiddenSurface(id, surface string) func(*contract) {
	return mutateChannel(id, func(item *channel) {
		item.ForbiddenSurfaces = removeString(item.ForbiddenSurfaces, surface)
	})
}

func setChannelRuntimeRole(id, role string) func(*contract) {
	return mutateChannel(id, func(item *channel) { item.RuntimeRole = role })
}

func setChannelBackgroundRule(id, rule string) func(*contract) {
	return mutateChannel(id, func(item *channel) { item.BackgroundRule = rule })
}

func setChannelUpdateMechanism(id, mechanism string) func(*contract) {
	return mutateChannel(id, func(item *channel) { item.UpdateMechanism = mechanism })
}

func mutateChannel(id string, mutate func(*channel)) func(*contract) {
	return func(value *contract) {
		for i := range value.Channels {
			if value.Channels[i].ID == id {
				mutate(&value.Channels[i])
			}
		}
	}
}
