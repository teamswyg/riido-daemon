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
	result.StoreArtifactRoots = append(result.StoreArtifactRoots, loaded.StoreArtifactRoots...)
	for _, item := range loaded.Channels {
		result.Channels = append(result.Channels, item.ID)
	}
	sort.Strings(result.Channels)
}
