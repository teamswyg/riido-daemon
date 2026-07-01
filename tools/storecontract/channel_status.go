package main

import "fmt"

const (
	channelStatusPreferredFirst   = "preferred-first"
	channelStatusRequiresRedesign = "requires-redesign"
	channelStatusStoreReviewReady = "store-review-ready"
)

func validateChannelStatus(item channel) []string {
	if item.ID == "msix-store" && item.Status != channelStatusStoreReviewReady {
		return []string{fmt.Sprintf(
			`channel "msix-store" status must be %s`,
			channelStatusStoreReviewReady,
		)}
	}
	switch item.Status {
	case channelStatusPreferredFirst,
		channelStatusRequiresRedesign,
		channelStatusStoreReviewReady:
		return nil
	default:
		return []string{fmt.Sprintf("channel %q status %q is not recognized", item.ID, item.Status)}
	}
}
