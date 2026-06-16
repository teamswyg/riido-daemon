package mwsdbridge

type Method string

const (
	MethodStatus        Method = "status"
	MethodGraph         Method = "graph"
	MethodDomain        Method = "domain"
	MethodHarness       Method = "harness"
	MethodOrchestration Method = "orchestration"
	MethodProjects      Method = "projects"
)
