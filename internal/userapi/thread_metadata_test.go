package userapi

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func threadMetadataFixture() string {
	raw := strings.Replace(threadHistoryFixture, `"thread_id":"thread-one"`, `"thread_id":"thread-one","requested_by":"human-requester","confirmation_reason":"needs_confirmation"`, 1)
	return strings.Replace(raw, `"message_id":"reply-one"`, `"message_id":"reply-one","reactions":[{"emoji":"1f44d","count":1,"actors":[{"kind":"human","id":"reactor-one"}]}]`, 1)
}

func TestThreadsPreserveRequesterConfirmationAndReactions(t *testing.T) {
	r := ThreadReader{Client: &threadFixtureClient{body: threadMetadataFixture(), status: 200}}
	list, err := r.List(t.Context(), "workspace-one", "task-one")
	if err != nil {
		t.Fatal("current CP metadata must not reject retained history", err)
	}
	show, err := r.Show(t.Context(), "workspace-one", "task-one", "thread-one")
	if err != nil {
		t.Fatal("current CP metadata must not reject a retained record", err)
	}
	for name, value := range map[string]any{"summary": list.Threads[0], "record": show.Thread} {
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]any
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		if fields["requested_by"] != "human-requester" || fields["confirmation_reason"] != "needs_confirmation" {
			t.Fatal(name, "lost requester or confirmation metadata")
		}
	}
	raw, err := json.Marshal(show.Thread.Messages[1])
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["reactions"]) != `[{"emoji":"1f44d","count":1,"actors":[{"kind":"human","id":"reactor-one"}]}]` {
		t.Fatal("message reactions must survive the CLI projection")
	}
}

func TestThreadsRejectAmbiguousNewMetadata(t *testing.T) {
	for name, raw := range map[string]string{
		"null requester":       strings.Replace(threadMetadataFixture(), `"requested_by":"human-requester"`, `"requested_by":null`, 1),
		"requester case alias": strings.Replace(threadMetadataFixture(), `"requested_by":`, `"Requested_by":`, 1),
		"unknown requester":    strings.Replace(threadMetadataFixture(), `"requested_by":`, `"requester_secret":`, 1),
		"null confirmation":    strings.Replace(threadMetadataFixture(), `"confirmation_reason":"needs_confirmation"`, `"confirmation_reason":null`, 1),
		"null reactions":       strings.Replace(threadHistoryFixture, `"message_id":"reply-one"`, `"message_id":"reply-one","reactions":null`, 1),
		"null actors":          strings.Replace(threadMetadataFixture(), `"actors":[{"kind":"human","id":"reactor-one"}]`, `"actors":null`, 1),
		"actor case alias":     strings.Replace(threadMetadataFixture(), `"kind":"human"`, `"Kind":"human"`, 1),
		"unknown actor field":  strings.Replace(threadMetadataFixture(), `"kind":"human"`, `"kind":"human","private":"secret"`, 1),
		"count wrong type":     strings.Replace(threadMetadataFixture(), `"count":1`, `"count":"1"`, 1),
		"duplicate actor ID":   strings.Replace(threadMetadataFixture(), `"id":"reactor-one"`, `"id":"reactor-one","id":"reactor-two"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			r := ThreadReader{Client: &threadFixtureClient{body: raw, status: 200}}
			if got, err := r.List(t.Context(), "workspace-one", "task-one"); got != nil || !errors.Is(err, ErrInvalidResponse) {
				t.Fatal("ambiguous metadata accepted", err)
			}
		})
	}
}
