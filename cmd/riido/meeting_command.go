package main

import (
	"strconv"
	"strings"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

type meetingCommand string

const (
	meetingCommandList meetingCommand = "list"
	meetingCommandGet  meetingCommand = "get"
	meetingCommandBody meetingCommand = "body"
)

func runMeeting(c contextIO, args []string) error {
	if len(args) == 0 || isHelpArg(args[0]) {
		printMeetingUsage(c.stderr)
		if len(args) == 0 {
			return c.fail(false, contextcli.NewRequestID(), contextcli.Fail(contextcli.CodeUsage, "meeting 하위 명령이 필요합니다", "riido meeting --help"))
		}
		return nil
	}
	switch meetingCommand(args[0]) {
	case meetingCommandList:
		// List selects work and budget with the same defaults as context get.
		call := contextGetCall()
		call.command, call.usage = "meeting list", printMeetingUsage
		return c.runContextRead(call, args[1:])
	case meetingCommandGet:
		if c.environment().Mode != contextcli.ModeAgent {
			return c.runUserMeeting(args[1:])
		}
		return c.runContextRead(meetingGetCall(), args[1:])
	case meetingCommandBody:
		return c.runUserResourceBody(args[1:], "meeting")
	default:
		return c.fail(wantsJSON(args), contextcli.NewRequestID(), contextcli.Failf(contextcli.CodeUsage, "알 수 없는 meeting 하위 명령: %s", args[0]))
	}
}

func meetingGetCall() contextReadCall {
	return contextReadCall{
		command: "meeting get",
		spec:    contextcli.ArgSpec{MaxPositionals: 1, Values: []string{"--budget"}},
		usage:   printMeetingUsage,
		payload: func(args contextcli.Args) (map[string]string, *contextcli.CLIError) {
			if len(args.Positionals) == 0 || strings.TrimSpace(args.Positionals[0]) == "" {
				return nil, contextcli.Fail(contextcli.CodeUsage, "미팅 ID를 지정해야 합니다", "riido meeting list --json 으로 ID를 확인하세요")
			}
			payload := map[string]string{"meeting": strings.TrimSpace(args.Positionals[0])}
			if budget := args.Value("--budget"); budget != "" {
				chars, err := strconv.Atoi(budget)
				if err != nil || chars < 1 {
					return nil, contextcli.Failf(contextcli.CodeUsage, "--budget 은 1 이상의 문자 수여야 합니다: %s", budget)
				}
				payload["budget_chars"] = strconv.Itoa(chars)
			}
			return payload, nil
		},
	}
}
