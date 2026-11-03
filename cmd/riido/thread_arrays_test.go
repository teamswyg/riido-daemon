package main

import (
	"encoding/json"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestThreadCLIRequiredArrays(t *testing.T) {
	for _, field := range []string{"messages", "assigned_agent_ids"} {
		for _, mode := range []string{"null", "empty"} {
			for _, command := range []string{"list", "show"} {
				t.Run(field+"/"+mode+"/"+command, func(t *testing.T) {
					var fixture map[string]any
					if err := json.Unmarshal([]byte(threadCLIHistory), &fixture); err != nil {
						t.Fatal(err)
					}
					var value any
					if mode == "empty" {
						value = []any{}
					}
					if field == "messages" {
						fixture["threads"].([]any)[0].(map[string]any)[field] = value
					} else {
						fixture[field] = value
					}
					raw, err := json.Marshal(fixture)
					if err != nil {
						t.Fatal(err)
					}
					s, _ := threadCLIService(t, string(raw), 200)
					h := newContextHarness(t, nil)
					h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return s, nil }
					args := []string{command}
					if command == "show" {
						args = append(args, "thread-one")
					}
					args = append(args, "--work", "task-one", "--workspace", "workspace-one", "--json")
					err = runThread(h.io, args)
					env := decodeEnvelope(t, h.stdout.Bytes())
					if mode == "null" {
						if err == nil || env.Error == nil || env.Data != nil {
							t.Fatal("null required array exposed a success projection")
						}
					} else if err != nil || env.Error != nil || env.Data == nil {
						t.Fatal("normal empty array rejected", err)
					}
				})
			}
		}
	}
}
