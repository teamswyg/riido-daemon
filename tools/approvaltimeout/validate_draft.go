package main

func mappedDraftCheck(event ApprovalEvent, mapped []MappedEvent) ManifestCheck {
	check := ManifestCheck{Name: "approval_event_maps_to_provider_draft", Expected: event.EventType}
	for _, row := range mapped {
		if row.EventKind == event.EventKind {
			check.Actual = row.EventType
			check.Pass = row.EventType == event.EventType
			return check
		}
	}
	return check
}

func skippedTimeoutCheck(kind string, skipped []SkippedEvent) ManifestCheck {
	check := ManifestCheck{Name: "timeout_not_provider_draft_owned", Expected: kind}
	for _, row := range skipped {
		if row.EventKind == kind {
			check.Actual, check.Pass = row.EventKind, true
			return check
		}
	}
	return check
}
