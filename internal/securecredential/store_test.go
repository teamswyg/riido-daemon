package securecredential

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestSaveVerifiesBeforeSendingCredential(t *testing.T) {
	called := false
	s := testStore(func(context.Context, string, []byte) ([]byte, error) { called = true; return nil, nil })
	s.verify = func(context.Context, string) error { return errors.New("untrusted helper") }
	if err := s.Save(context.Background(), []byte("private-value")); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error=%v", err)
	}
	if called {
		t.Fatal("credential was sent before signer verification")
	}
}

func TestLoadSanitizesHelperErrors(t *testing.T) {
	s := testStore(func(context.Context, string, []byte) ([]byte, error) {
		return []byte(`{"version":1,"ok":false,"error":{"code":"INTERACTION_REQUIRED","message":"secret-from-provider"}}`), nil
	})
	_, err := s.Load(context.Background())
	if !errors.Is(err, ErrInteractionRequired) || strings.Contains(err.Error(), "secret-from-provider") {
		t.Fatalf("error=%v", err)
	}
}

func TestLoadAndSaveOpaqueCredential(t *testing.T) {
	var sent request
	s := testStore(func(_ context.Context, _ string, body []byte) ([]byte, error) {
		if err := json.Unmarshal(body, &sent); err != nil {
			return nil, err
		}
		if sent.Operation == "load" {
			return []byte(`{"version":1,"ok":true,"data":{"present":true,"value":"cHJpdmF0ZS12YWx1ZQ=="}}`), nil
		}
		return []byte(`{"version":1,"ok":true,"data":{"present":true}}`), nil
	})
	value, err := s.Load(context.Background())
	if err != nil || string(value) != "private-value" {
		t.Fatalf("load failed: %v", err)
	}
	if err = s.Save(context.Background(), value); err != nil {
		t.Fatal(err)
	}
	if sent.Value != "cHJpdmF0ZS12YWx1ZQ==" || !sent.Interactive || sent.Key != "account-key-0123456789" {
		t.Fatal("request contract differs")
	}
}

func TestRejectsUnboundedAndMalformedResponse(t *testing.T) {
	for _, response := range []string{strings.Repeat("x", maxMessageBytes+1), `{"version":2,"ok":true}`, `{"version":1,"ok":true,"data":{"present":true,"value":"!"}}`, "{} trailing"} {
		s := testStore(func(context.Context, string, []byte) ([]byte, error) { return []byte(response), nil })
		if _, err := s.Load(context.Background()); !errors.Is(err, ErrUnavailable) {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestSaveRejectsOversizedValueWithoutHelper(t *testing.T) {
	called := false
	s := testStore(func(context.Context, string, []byte) ([]byte, error) { called = true; return nil, nil })
	if err := s.Save(context.Background(), make([]byte, maxValueBytes+1)); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error=%v", err)
	}
	if called {
		t.Fatal("oversized value sent")
	}
}

func testStore(run func(context.Context, string, []byte) ([]byte, error)) *Store {
	return &Store{HelperPath: "/trusted/helper", Key: "account-key-0123456789", Interactive: true, verify: func(context.Context, string) error { return nil }, run: run}
}

func TestHelperSuccessRequiresOperationEvidence(t *testing.T) {
	for _, body := range []string{`{"version":1,"ok":true}`, `{"version":1,"ok":true,"data":{}}`, `{"version":1,"ok":true,"data":{"present":false}}`, `{"version":1,"ok":true,"data":{"present":true,"value":"unexpected"}}`} {
		s := testStore(func(context.Context, string, []byte) ([]byte, error) { return []byte(body), nil })
		if err := s.Save(context.Background(), []byte("fixture")); !errors.Is(err, ErrUnavailable) {
			t.Fatalf("save accepted invalid evidence: %v", err)
		}
	}
}

func TestShortKeysNeverReachHelper(t *testing.T) {
	called := false
	s := testStore(func(context.Context, string, []byte) ([]byte, error) { called = true; return nil, nil })
	s.Key = "short"
	_, _ = s.Status(context.Background())
	if called {
		t.Fatal("invalid helper key was sent")
	}
}
