package securecredential

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
)

type fakeSession struct {
	trips  int
	closed bool
	fail   bool
}

func (f *fakeSession) roundTrip(_ context.Context, _ []byte) ([]byte, error) {
	f.trips++
	if f.fail {
		return nil, errors.New("broken")
	}
	value := base64.StdEncoding.EncodeToString([]byte("fixture"))
	return []byte(`{"version":1,"ok":true,"data":{"present":true,"value":"` + value + `"}}` + "\n"), nil
}
func (f *fakeSession) close() { f.closed = true }

func withFakeSessions(t *testing.T) *[]*fakeSession {
	t.Helper()
	started := []*fakeSession{}
	previousStart, previousSupported := startSession, sessionSupported
	startSession = func(context.Context, string) (sessionConn, error) {
		conn := &fakeSession{}
		started = append(started, conn)
		return conn, nil
	}
	sessionSupported = true
	CloseSessions()
	t.Cleanup(func() { CloseSessions(); startSession, sessionSupported = previousStart, previousSupported })
	return &started
}

func sessionStore(shared, interactive bool) *Store {
	return &Store{
		HelperPath: "/fixture/helper", Key: "fixture_account_key_5900", Interactive: interactive, SharedSession: shared,
		verify: func(context.Context, string) error { return nil },
	}
}

func TestSharedSessionStartsOneHelperPerCommand(t *testing.T) {
	started := withFakeSessions(t)
	credential, repair := sessionStore(true, true), sessionStore(true, true)
	repair.Key = "fixture_repair_key_5900"
	for _, store := range []*Store{credential, credential, repair} {
		if value, err := store.Load(context.Background()); err != nil || string(value) != "fixture" {
			t.Fatalf("load = %q, %v", value, err)
		}
	}
	if len(*started) != 1 || (*started)[0].trips != 3 {
		t.Fatalf("sessions = %d, trips = %d; want one session serving every request", len(*started), (*started)[0].trips)
	}
}

func TestNoninteractiveAndUnsharedStoresNeverUseSession(t *testing.T) {
	started := withFakeSessions(t)
	for _, store := range []*Store{sessionStore(true, false), sessionStore(false, true)} {
		store.run = func(context.Context, string, []byte) ([]byte, error) {
			return []byte(`{"version":1,"ok":true,"data":{"present":false}}`), nil
		}
		if present, err := store.Status(context.Background()); err != nil || present {
			t.Fatalf("status = %v, %v", present, err)
		}
	}
	if len(*started) != 0 {
		t.Fatalf("started %d sessions for per-request stores", len(*started))
	}
}

func TestBrokenSessionIsDroppedAndNotReused(t *testing.T) {
	started := withFakeSessions(t)
	store := sessionStore(true, true)
	if _, err := store.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	(*started)[0].fail = true
	if _, err := store.Load(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("broken session error = %v", err)
	}
	if !(*started)[0].closed {
		t.Fatal("broken session stayed open")
	}
	if _, err := store.Load(context.Background()); err != nil || len(*started) != 2 {
		t.Fatalf("restart = %v, sessions %d", err, len(*started))
	}
}
