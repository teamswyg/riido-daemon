package main

func taskFixtureScenarios(fixture taskFixture) []scenario {
	var out []scenario
	if fixture.Team.ID != "" {
		out = append(out, fixture.Team)
	}
	if fixture.Create.ID != "" {
		out = append(out, fixture.Create)
	}
	return out
}

func taskFixtureBlocked(fixture taskFixture) bool {
	return fixture.Team.ID != "" && fixture.Team.Status != statusPassed ||
		fixture.Create.ID != "" && fixture.Create.Status != statusPassed
}

func finishTaskFlow(cfg config, fixture taskFixture, out, tail []scenario) []scenario {
	out = append(out, tail...)
	if fixture.Created() {
		out = append(out, cleanupTaskFixture(cfg, fixture))
	}
	return out
}
