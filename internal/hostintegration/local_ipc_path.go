package hostintegration

func namedPipePath(channel DistributionChannel, owner LocalIPCOwner, name string) string {
	return `\\.\pipe\riido-` + string(channel) + "-" + string(owner) + "-" + name
}
