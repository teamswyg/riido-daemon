package main

type options struct {
	Repo              string
	Manifest          string
	CandidateIn       string
	EvidenceOut       string
	WriteDoc          bool
	CheckDoc          bool
	GitHubAnnotations bool
}
