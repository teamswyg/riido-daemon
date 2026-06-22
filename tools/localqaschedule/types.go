package main

type config struct {
	repo             *string
	s3Prefix         *string
	evidenceOut      *string
	productEvidence  *string
	clientRoot       *string
	productBaseURL   *string
	productWorkspace *string
	productStorage   *string
	productTaskID    *string
	productAgentID1  *string
	productAgentID2  *string
	productComment   *string
	taskMutations    *bool
	startClient      *bool
	runProduct       *bool
	label            *string
	plistPath        *string
	hour             *int
	minute           *int
	install          *bool
	runAtLoad        *bool
}

type schedulePaths struct {
	repo      string
	plist     string
	stdout    string
	stderr    string
	launchctl string
}
