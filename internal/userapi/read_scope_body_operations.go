package userapi

const readBacklogBodyQuery = `query RiidoCLIBacklogBody($backlogId:String!,$teamId:String!,$part:String!,$version:String!,$offset:Int!,$limit:Int!,$selector:RiidoReadPartSelectorInput){
 getBacklogReadScopeBody(backlogId:$backlogId,teamId:$teamId,part:$part,version:$version,offset:$offset,limit:$limit,selector:$selector){` + readBodyFields + `}
}`

const readIntakeBodyQuery = `query RiidoCLIIntakeBody($intakeId:String!,$teamId:String!,$part:String!,$version:String!,$offset:Int!,$limit:Int!,$selector:RiidoReadPartSelectorInput){
 getIntakeReadScopeBody(intakeId:$intakeId,teamId:$teamId,part:$part,version:$version,offset:$offset,limit:$limit,selector:$selector){` + readBodyFields + `}
}`

const readMeetingBodyQuery = `query RiidoCLIMeetingBody($meetingId:String!,$teamId:String!,$part:String!,$version:String!,$offset:Int!,$limit:Int!,$selector:RiidoReadPartSelectorInput){
 getMeetingReadScopeBody(meetingId:$meetingId,teamId:$teamId,part:$part,version:$version,offset:$offset,limit:$limit,selector:$selector){` + readBodyFields + `}
}`

func validBodyResourceKind(kind string) bool {
	return kind == "" || kind == "component" || kind == "backlog" || kind == "intake" || kind == "meeting"
}

func readBodyOperation(kind string) (query, field, idKey string) {
	switch kind {
	case "backlog":
		return readBacklogBodyQuery, "getBacklogReadScopeBody", "backlogId"
	case "intake":
		return readIntakeBodyQuery, "getIntakeReadScopeBody", "intakeId"
	case "meeting":
		return readMeetingBodyQuery, "getMeetingReadScopeBody", "meetingId"
	default:
		return readScopeBodyQuery, "getComponentReadScopeBody", "componentId"
	}
}
