package main

type docLink struct {
	Title string
	File  string
}

var architectureLinks = []docLink{
	{"Distribution decisions", "decisions.md"},
	{"Target matrix", "target-matrix.md"},
	{"MSIX acceptance criteria", "msix-acceptance.md"},
	{"Mac App Store acceptance criteria", "mac-app-store-acceptance.md"},
	{"Package boundaries", "package-boundaries.md"},
	{"macOS helper / login item strategy", "macos-helper-login.md"},
	{"Windows MSIX runtime / packaging strategy", "windows-msix-runtime.md"},
}

var daemonLinks = []docLink{
	{"Required daemon changes", "required-daemon-changes.md"},
	{"Required server changes", "required-server-changes.md"},
	{"Review notes contract", "review-notes-contract.md"},
	{"Executable contract", "executable-contract.md"},
	{"External sources", "external-sources.md"},
}
