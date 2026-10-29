package main

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// A cursor is issued by the captured manifest. The CLI cannot derive a page
// from a source ID, switch to live reads, or substitute a user credential.
func contextPartCall() contextReadCall {
	return contextReadCall{
		command: "context part",
		spec:    contextcli.ArgSpec{MaxPositionals: 1},
		usage:   printContextGroupUsage,
		payload: func(args contextcli.Args) (map[string]string, *contextcli.CLIError) {
			if len(args.Positionals) != 1 || strings.TrimSpace(args.Cursor) == "" {
				return nil, contextcli.Fail(contextcli.CodeUsage, "자료 ID와 저장된 페이지의 --cursor가 필요합니다", "riido context part <SOURCE-ID> --cursor <CURSOR>")
			}
			ref, err := contextcli.ParseWorkRef(args.Positionals[0])
			if err != nil {
				return nil, err
			}
			if ref.Kind != contextcli.WorkRefID {
				return nil, contextcli.Fail(contextcli.CodeUsage, "페이지에 표시된 실제 자료 ID를 사용하세요", "표시 key 대신 원 응답의 source ID를 지정하세요")
			}
			return map[string]string{"work": ref.Value}, nil
		},
	}
}
