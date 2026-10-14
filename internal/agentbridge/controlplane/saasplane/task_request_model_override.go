package saasplane

import (
	assignmentcontract "github.com/teamswyg/riido-contracts/assignment"
	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
)

func assignmentProviderModelOverride(assignment assignmentcontract.Assignment) string {
	return providercatalog.ModelOverride(assignment.RuntimeProvider, assignment.ModelID)
}
