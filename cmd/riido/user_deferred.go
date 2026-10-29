package main

import (
	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userapi"
)

func deferredContentWarning() contextcli.Warning {
	return contextcli.Warning{Code: "CONTENT_DEFERRED", Message: "contentPage가 있는 빈 항목은 실제 빈 자료가 아닙니다. riido work body 또는 riido meeting body에 원 part·version·selector를 전달하고 nextOffset 순서로 이어 읽으세요. --limit은 바이트 수, --comments-limit·--board-limit은 원 페이지 항목 수입니다. 완성된 content를 합쳐 JSON으로 읽은 뒤 그 안의 다음 페이지를 따르세요"}
}

func scopeHasDeferredContent(scope userapi.ReadScope) bool {
	for _, work := range scope.Works {
		if work.ContentPage != nil || work.Comments.ContentPage != nil {
			return true
		}
	}
	for _, meeting := range scope.Meetings {
		if meeting.Snapshot != nil && meetingHasDeferredContent(*meeting.Snapshot) {
			return true
		}
	}
	return false
}

func meetingHasDeferredContent(meeting userapi.ReadMeeting) bool {
	return meeting.ContentPage != nil || (meeting.Board != nil && meeting.Board.ContentPage != nil)
}
