package main

type packageRole struct {
	Label    string   `json:"label"`
	Packages []string `json:"packages"`
	Role     string   `json:"role"`
}

type importRule struct {
	Group             string   `json:"group"`
	PackagePrefixes   []string `json:"package_prefixes"`
	MayImport         string   `json:"may_import"`
	ForbiddenPrefixes []string `json:"forbidden_prefixes"`
	MustNotImport     string   `json:"must_not_import"`
}

type port struct {
	Port     string `json:"port"`
	Package  string `json:"package"`
	Adapters string `json:"adapters"`
}

type factorBoundary struct {
	Configuration string `json:"configuration"`
	TestGates     string `json:"test_gates"`
	State         string `json:"state"`
	Listener      string `json:"listener"`
}
