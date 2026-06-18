package main

func requiredChannelSet() map[string]bool {
	return map[string]bool{
		"developer-id":  false,
		"mac-app-store": false,
		"msix-sideload": false,
		"msix-store":    false,
	}
}

func markRequiredChannel(required map[string]bool, id string) {
	if _, ok := required[id]; ok {
		required[id] = true
	}
}
