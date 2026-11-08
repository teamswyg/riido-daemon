package main

import (
	"net/http"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/userauth"
)

func TestAlreadyLinkedDoesNotRecommendRepairOrRetry(t *testing.T) {
	err := userLinkError(&userauth.DeviceLinkError{Code: "LINK_ALREADY_LINKED", StatusCode: http.StatusConflict})
	if err.Code != contextcli.CodeConflict || !err.NoRetry || err.NextAction != "재로그인이나 repair로 연결 기기가 바뀌지 않습니다. 기존 기기 연결을 확인하세요" {
		t.Fatal("wrong existing-link guidance", err)
	}
	legacy := userLinkError(&userauth.DeviceLinkError{Code: "LINK_REQUEST_REJECTED", StatusCode: http.StatusConflict})
	if legacy.NextAction == err.NextAction {
		t.Fatal("ambiguous conflict became existing-link evidence")
	}
}
