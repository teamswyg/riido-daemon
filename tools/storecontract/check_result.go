package main

import "sort"

func newCheckResult(contractPath string) checkResult {
	return checkResult{
		SchemaVersion: checkSchemaVersion,
		ContractPath:  contractPath,
		Status:        "failed",
	}
}

func (result *checkResult) addContractMetadata(loaded contract) {
	result.Product = loaded.Product
	result.ChannelStatuses = map[string]string{}
	result.StoreArtifactRoots = append(result.StoreArtifactRoots, loaded.StoreArtifactRoots...)
	for _, item := range loaded.Channels {
		result.Channels = append(result.Channels, item.ID)
		result.ChannelStatuses[item.ID] = item.Status
	}
	sort.Strings(result.Channels)
}
