package hostintegration

import "strings"

func helperRuntimeEndpoint(in HelperRuntimePlanInput) (LocalIPCEndpoint, error) {
	name := strings.TrimSpace(in.EndpointName)
	if name == "" {
		name = "riido"
	}
	return DefaultLocalIPCEndpoint(LocalIPCEndpointInput{
		Channel:     in.Channel,
		HostOS:      in.HostOS,
		AppDataRoot: in.AppDataRoot,
		Owner:       LocalIPCOwnerHelper,
		Name:        name,
	})
}
