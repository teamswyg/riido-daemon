package userapi

// Fixed v2 selections are opt-in; v1 query bytes remain unchanged.
const readBodyV2Fields = `schemaVersion workspaceId teamId resourceKind ` + readBodyFields

const readScopeCommentsV2Fields = `schemaVersion workspaceId teamId contentPage{` + readBodyV2Fields + `} resourceId resourceKind parentCommentId limit version status page nextPage complete items{id status createdAt updatedAt body replyCount
 author{id status workspaceProfile{nickname status avatar{fileLocation}}} workspaceAuthor{id nickname status}
 reactions{emoji count authorIds} isFromSlack slackMessageUrl isFromDiscord discordMessageUrl notificationSubscribed}`

const readScopeV2Query = `query RiidoCLIReadScope($componentId:String!,$teamId:String,$workspaceId:String!){
 getComponentReadScope(componentId:$componentId,teamId:$teamId,workspaceId:$workspaceId,locatorVersion:2){
  schemaVersion workspaceId rootId teamId status selection{relatedWorkLimit relatedMeetingLimit childrenLimit}
  works{id kind workspaceId teamId summary{id kind title updatedAt keyNumber componentType branchName} contentPage{` + readBodyV2Fields + `} reasons{kind relation source}
   details{id componentType title keyNumber branchName startDate endDate estimate totalPriorityScore
    isTemplate isPublished isDone doneDate cancelledAt cancellationReason isArchived archivedDate
    isOnboardingComponent createdAt updatedAt deletedAt notificationSubscribed estimateMethod
    createdBy{id status workspaceProfile{nickname status avatar{fileLocation}}}
    lastUpdatedBy{id status workspaceProfile{nickname status avatar{fileLocation}}} createdByWorkspaceMember{id nickname status} lastUpdatedByWorkspaceMember{id nickname status}
    status{id statusName statusType statusOrder statusColor colorPalette{id colorHex colorHexLight textColorHex textColorHexLight statusColorHex statusColorHexLight}}
    componentMembers{id memberOrder workspaceMember{id nickname status} teamMember{id status workspaceProfile{nickname status avatar{fileLocation}}}}
    componentLabels{id labelOrder label{id labelOrder labelName labelColor colorPalette{id colorHex colorHexLight textColorHex textColorHexLight statusColorHex statusColorHexLight}}}
    task{id componentType keyNumber title} milestone{id componentType keyNumber title} project{id componentType keyNumber title}}
   backlogDetails{id title importance estimate expectedStartingDate urgent backlogOrder movedDate reviewState reviewedAt cancellationReason createdAt updatedAt deletedAt
    section{id sectionName} assignee{` + readScopeMemberFields + `} createdBy{` + readScopeMemberFields + `} lastUpdatedBy{` + readScopeMemberFields + `} createdByWorkspaceMember{id nickname status} lastUpdatedByWorkspaceMember{id nickname status}
    backlogLabels{id labelOrder label{id labelName labelColor colorPalette{id colorHex colorHexLight}}}
    voters{` + readScopeMemberFields + `} notificationSubscribed isNew}
   intakeDetails{id title status urgent endDate closedDate createdAt updatedAt deletedAt isPermaDeleted estimate estimateMethod
    intakeLabels{id labelOrder label{id labelName labelColor colorPalette{id colorHex colorHexLight}}}
    assignee{` + readScopeMemberFields + `} createdBy{` + readScopeMemberFields + `} lastUpdatedBy{` + readScopeMemberFields + `} createdByWorkspaceMember{id nickname status} lastUpdatedByWorkspaceMember{id nickname status} notificationSubscribed}
   body{` + readBodyV2Fields + `} descriptionBody{` + readBodyV2Fields + `} alternateBody{` + readBodyV2Fields + `}
   comments{` + readScopeCommentsV2Fields + `}
   links{id title url description createdAt logo} github{id title number state updatedAt commentsCount aiShortSummary htmlUrl body lastSyncedAt aiSummary aiLastSyncedAt comments{id body authorName} timelineEventTypes}
   repositories{id fullName isPrivate repositoryUrl source} sprints{id number title startDate endDate} origin{id kind title isDeleted} figma{id url fileKey nodeId}
   priorities{id score priority{id priorityName allocationRatio priorityOrder}}
   checklists{id groupName groupOrder items{id itemName itemOrder isChecked}}
   attachments{id createdAt fileName fileSize mimeType fileLocation uploader{` + readScopeMemberFields + `} workspaceUploader{id nickname status}}
   partsStatus{name status complete}}
  meetings{id teamId reasons{kind relation source}}
 }
}`

const readScopeBodyV2Query = `query RiidoCLIReadScopeBody($componentId:String!,$workspaceId:String!,$teamId:String,$part:String!,$version:String!,$offset:Int!,$limit:Int!,$selector:RiidoReadPartSelectorInput){
 getComponentReadScopeBody(componentId:$componentId,workspaceId:$workspaceId,teamId:$teamId,locatorVersion:2,part:$part,version:$version,offset:$offset,limit:$limit,selector:$selector){` + readBodyV2Fields + `}
}`
const readBacklogBodyV2Query = `query RiidoCLIBacklogBody($backlogId:String!,$workspaceId:String!,$teamId:String,$part:String!,$version:String!,$offset:Int!,$limit:Int!,$selector:RiidoReadPartSelectorInput){
 getBacklogReadScopeBody(backlogId:$backlogId,workspaceId:$workspaceId,teamId:$teamId,locatorVersion:2,part:$part,version:$version,offset:$offset,limit:$limit,selector:$selector){` + readBodyV2Fields + `}
}`
const readIntakeBodyV2Query = `query RiidoCLIIntakeBody($intakeId:String!,$workspaceId:String!,$teamId:String,$part:String!,$version:String!,$offset:Int!,$limit:Int!,$selector:RiidoReadPartSelectorInput){
 getIntakeReadScopeBody(intakeId:$intakeId,workspaceId:$workspaceId,teamId:$teamId,locatorVersion:2,part:$part,version:$version,offset:$offset,limit:$limit,selector:$selector){` + readBodyV2Fields + `}
}`
const readScopeCommentsV2Query = `query RiidoCLIReadScopeComments($resourceId:String!,$resourceKind:String!,$workspaceId:String!,$teamId:String,$page:Int!,$limit:Int!,$parentCommentId:String,$version:String){
 getReadScopeComments(resourceId:$resourceId,resourceKind:$resourceKind,workspaceId:$workspaceId,teamId:$teamId,locatorVersion:2,page:$page,limit:$limit,parentCommentId:$parentCommentId,version:$version){` + readScopeCommentsV2Fields + `}
}`

const workV2Query = `query RiidoCLIWork($componentId:String!,$teamId:String,$workspaceId:String!){getComponent(componentId:$componentId,teamId:$teamId,workspaceId:$workspaceId,locatorVersion:2,getComponentDto:{getDocument:"true"}){schemaVersion workspaceId ownerTeamId canEdit createdByWorkspaceMember{id nickname status} lastUpdatedByWorkspaceMember{id nickname status} id componentType title keyNumber description createdAt updatedAt document{id HTMLContent tiptapDocumentId} startDate endDate estimate totalPriorityScore isDone isArchived status{id statusName statusType} componentMembers{id memberOrder workspaceMember{id nickname status} teamMember{id status workspaceProfile{nickname status}}} componentLabels{id labelOrder label{id labelName}}}}`
