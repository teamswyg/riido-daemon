package main

func validateContractShape(loaded contract) []string {
	var problems []string
	problems = append(problems, validateContractIdentity(loaded)...)
	problems = append(problems, validateProviderCLINames(loaded.ExternalProviderCLINames)...)
	problems = append(problems, validateContractCollections(loaded)...)
	return append(problems, validateChannels(loaded.Channels)...)
}
