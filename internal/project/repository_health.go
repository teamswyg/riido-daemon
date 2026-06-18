package project

type RepositoryHealth string

const (
	RepositoryReady          RepositoryHealth = "ready"
	RepositoryMissingLocal   RepositoryHealth = "missing-local"
	RepositoryMissingGit     RepositoryHealth = "missing-git"
	RepositoryRemoteMismatch RepositoryHealth = "remote-mismatch"
)
