package userapi

const readBodyFields = `resourceId part status source format version offset content nextOffset complete selector{parentCommentId page limit boardCursor}`

const readScopeQuery = `query RiidoCLIReadScope($componentId:String!,$teamId:String!){
 getComponentReadScope(componentId:$componentId,teamId:$teamId){
  schemaVersion rootId teamId status selection{relatedWorkLimit relatedMeetingLimit childrenLimit}
  works{id kind teamId summary{id kind title updatedAt keyNumber componentType branchName} contentPage{` + readBodyFields + `} reasons{kind relation source}
   details{id componentType title keyNumber branchName startDate endDate estimate totalPriorityScore
    isTemplate isPublished isDone doneDate cancelledAt cancellationReason isArchived archivedDate
    isOnboardingComponent createdAt updatedAt deletedAt notificationSubscribed estimateMethod
    createdBy{id status workspaceProfile{nickname status avatar{fileLocation}}}
    lastUpdatedBy{id status workspaceProfile{nickname status avatar{fileLocation}}}
    status{id statusName statusType statusOrder statusColor colorPalette{id colorHex colorHexLight textColorHex textColorHexLight statusColorHex statusColorHexLight}}
    componentMembers{id memberOrder teamMember{id status workspaceProfile{nickname status avatar{fileLocation}}}}
    componentLabels{id labelOrder label{id labelOrder labelName labelColor colorPalette{id colorHex colorHexLight textColorHex textColorHexLight statusColorHex statusColorHexLight}}}
    task{id componentType keyNumber title} milestone{id componentType keyNumber title} project{id componentType keyNumber title}}
   backlogDetails{id title importance estimate expectedStartingDate urgent backlogOrder movedDate reviewState reviewedAt cancellationReason createdAt updatedAt deletedAt
    section{id sectionName} assignee{` + readScopeMemberFields + `} createdBy{` + readScopeMemberFields + `} lastUpdatedBy{` + readScopeMemberFields + `}
    backlogLabels{id labelOrder label{id labelName labelColor colorPalette{id colorHex colorHexLight}}}
    voters{` + readScopeMemberFields + `} notificationSubscribed isNew}
   intakeDetails{id title status urgent endDate closedDate createdAt updatedAt deletedAt isPermaDeleted estimate estimateMethod
    intakeLabels{id labelOrder label{id labelName labelColor colorPalette{id colorHex colorHexLight}}}
    assignee{` + readScopeMemberFields + `} createdBy{` + readScopeMemberFields + `} lastUpdatedBy{` + readScopeMemberFields + `} notificationSubscribed}
   body{` + readBodyFields + `} descriptionBody{` + readBodyFields + `} alternateBody{` + readBodyFields + `}
   comments{` + readScopeCommentFields + `}
   links{id title url description createdAt logo} github{id title number state updatedAt commentsCount aiShortSummary htmlUrl body lastSyncedAt aiSummary aiLastSyncedAt comments{id body authorName} timelineEventTypes}
   repositories{id fullName isPrivate repositoryUrl source} sprints{id number title startDate endDate} origin{id kind title isDeleted} figma{id url fileKey nodeId}
   priorities{id score priority{id priorityName allocationRatio priorityOrder}}
   checklists{id groupName groupOrder items{id itemName itemOrder isChecked}}
   attachments{id createdAt fileName fileSize mimeType fileLocation uploader{` + readScopeMemberFields + `}}
   partsStatus{name status complete}}
  meetings{id teamId reasons{kind relation source}}
 }
}`

const readScopeMemberFields = `id status workspaceProfile{nickname status avatar{fileLocation}}`

const readScopeBodyQuery = `query RiidoCLIReadScopeBody($componentId:String!,$teamId:String!,$part:String!,$version:String!,$offset:Int!,$limit:Int!,$selector:RiidoReadPartSelectorInput){
 getComponentReadScopeBody(componentId:$componentId,teamId:$teamId,part:$part,version:$version,offset:$offset,limit:$limit,selector:$selector){` + readBodyFields + `}
}`
