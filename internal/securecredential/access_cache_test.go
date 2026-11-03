package securecredential

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestAccessCacheOperations(t *testing.T) {
	ops := []string{}
	s := testStore(func(_ context.Context, _ string, body []byte) ([]byte, error) {
		var sent request
		if err := json.Unmarshal(body, &sent); err != nil {
			return nil, err
		}
		ops = append(ops, sent.Operation)
		switch sent.Operation {
		case "access-load":
			return []byte(`{"version":1,"ok":true,"data":{"present":true,"value":"Z3JhbnQ="}}`), nil
		case "access-save":
			return []byte(`{"version":1,"ok":true,"data":{"present":true}}`), nil
		default:
			return []byte(`{"version":1,"ok":true,"data":{"present":false}}`), nil
		}
	})
	value, err := s.LoadAccess(context.Background())
	if err != nil || string(value) != "grant" {
		t.Fatalf("load %q %v", value, err)
	}
	if err = s.SaveAccess(context.Background(), []byte("grant")); err != nil {
		t.Fatal(err)
	}
	if err = s.DeleteAccess(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(ops) != 3 || ops[0] != "access-load" || ops[1] != "access-save" || ops[2] != "access-delete" {
		t.Fatalf("ops %v", ops)
	}
	if err = s.SaveAccess(context.Background(), make([]byte, maxAccessBytes+1)); !errors.Is(err, ErrUnavailable) || len(ops) != 3 {
		t.Fatal("oversized access grant reached the helper")
	}
}

func TestAccessCacheOlderHelperIsUnsupported(t *testing.T) {
	s := testStore(func(context.Context, string, []byte) ([]byte, error) {
		return []byte(`{"version":1,"ok":false,"error":{"code":"INVALID_REQUEST","message":"Invalid helper request."}}`), nil
	})
	if _, err := s.LoadAccess(context.Background()); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if err := s.DeleteAccess(context.Background()); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	notFound := testStore(func(context.Context, string, []byte) ([]byte, error) {
		return []byte(`{"version":1,"ok":false,"error":{"code":"NOT_FOUND","message":"Credential not found."}}`), nil
	})
	if _, err := notFound.LoadAccess(context.Background()); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	malformed := testStore(func(context.Context, string, []byte) ([]byte, error) {
		return []byte(`{"version":1,"ok":true,"data":{"present":false}}`), nil
	})
	if _, err := malformed.LoadAccess(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}
