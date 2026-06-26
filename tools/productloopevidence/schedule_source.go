package main

func loadQASchedule(root string, m manifest) (qaScheduleSource, error) {
	var schedule qaScheduleSource
	err := loadJSON(repoPath(root, m.LocalQAScheduleManifest), &schedule)
	return schedule, err
}
