package userapi

const readScopeCommentFields = `contentPage{` + readBodyFields + `} resourceId resourceKind parentCommentId limit version status page nextPage complete items{id status createdAt updatedAt body replyCount
 author{id status workspaceProfile{nickname status avatar{fileLocation}}}
 reactions{emoji count authorIds} isFromSlack slackMessageUrl isFromDiscord discordMessageUrl notificationSubscribed}`

const readScopeCommentsQuery = `query RiidoCLIReadScopeComments($resourceId:String!,$resourceKind:String!,$teamId:String!,$page:Int!,$limit:Int!,$parentCommentId:String,$version:String){
 getReadScopeComments(resourceId:$resourceId,resourceKind:$resourceKind,teamId:$teamId,page:$page,limit:$limit,parentCommentId:$parentCommentId,version:$version){` + readScopeCommentFields + `}
}`
