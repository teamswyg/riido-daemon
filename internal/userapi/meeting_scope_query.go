package userapi

const (
	meetingScopeMemberFields  = `id status workspaceProfile{nickname status isActivated avatar{fileLocation}}`
	meetingScopeActorFields   = `id status workspaceProfile{nickname status avatar{fileLocation}}`
	meetingScopePaletteFields = `id colorHex colorHexLight`
)

const meetingScopeQuery = `query RiidoCLIMeetingReadScope($meetingId:String!,$teamId:String!,$boardCursor:String,$boardLimit:Int){
 getMeetingReadScope(meetingId:$meetingId,teamId:$teamId,boardCursor:$boardCursor,boardLimit:$boardLimit){
  id teamId status summary{id kind title updatedAt keyNumber componentType branchName} contentPage{` + readBodyFields + `}
  details{id createdAt updatedAt title startTime endTime isAllDay meetingLocation meetingColor
   isTemplate templateStartTime templateEndTime colorPalette{` + meetingScopePaletteFields + `}
   meetingAttendees{id memberOrder teamMember{` + meetingScopeMemberFields + `}}
   meetingParticipants{id memberOrder teamMember{` + meetingScopeMemberFields + `}}
   createdBy{` + meetingScopeActorFields + `} lastUpdatedBy{` + meetingScopeActorFields + `}}
  labels{id labelName labelColor labelOrder colorPalette{` + meetingScopePaletteFields + `}
   labelGroup{id groupName groupColor colorPalette{` + meetingScopePaletteFields + `}}}
  board{contentPage{` + readBodyFields + `} version offset nextCursor complete
   groups{id meetingBoardGroupOrder meetingBoardGroup{id groupName groupColor groupOrder colorPalette{` + meetingScopePaletteFields + `}}}
   cards{id groupId body boardOrder teamMember{` + meetingScopeMemberFields + `}
    component{id componentType keyNumber title} conversedDate}}
  body{` + readBodyFields + `} descriptionBody{` + readBodyFields + `} alternateBody{` + readBodyFields + `}
  partsStatus{name status complete}
 }
}`
