package main

import (
	"fmt"
	"sort"
)

func validateChannels(channels []channel) []string {
	required := requiredChannelSet()
	var problems []string
	for _, item := range channels {
		markRequiredChannel(required, item.ID)
		problems = append(problems, validateChannelFields(item)...)
		problems = append(problems, validateCommonForbiddenSurfaces(item)...)
		problems = append(problems, validateChannelPolicy(item)...)
	}
	for id, seen := range required {
		if !seen {
			problems = append(problems, fmt.Sprintf("required channel %q is missing", id))
		}
	}
	sort.Strings(problems)
	return problems
}

func validateChannelPolicy(item channel) []string {
	var problems []string
	problems = append(problems, validateDeveloperIDSurfaces(item)...)
	problems = append(problems, validateMacAppStoreSurfaces(item)...)
	return append(problems, validateMSIXSurfaces(item)...)
}
