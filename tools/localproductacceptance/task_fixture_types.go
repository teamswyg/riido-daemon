package main

type taskFixture struct {
	TaskID  string
	TeamID  string
	Title   string
	Team    scenario
	Create  scenario
	Cleanup scenario
}

func (f taskFixture) Created() bool {
	return f.TaskID != "" && f.TeamID != ""
}
