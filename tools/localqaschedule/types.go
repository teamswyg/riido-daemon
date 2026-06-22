package main

type config struct {
	repo      *string
	s3Prefix  *string
	label     *string
	plistPath *string
	hour      *int
	minute    *int
	install   *bool
	runAtLoad *bool
}

type schedulePaths struct {
	repo      string
	plist     string
	stdout    string
	stderr    string
	launchctl string
}
