package workdir

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// A daemon has one owner per workdir root. Serialize read/compare/write across
// adapter instances as well, so a second pin cannot silently replace ownership.
var nativeSessionMu sync.Mutex

func nativeSessionPath(provider, sessionID string) string {
	hash := sha256.Sum256([]byte(provider + "\x00" + sessionID))
	return filepath.Join(".native-sessions", hex.EncodeToString(hash[:])+".json")
}

func (a *FSAdapter) LookupNativeSession(provider, sessionID string) (NativeSession, bool, error) {
	nativeSessionMu.Lock()
	defer nativeSessionMu.Unlock()
	return a.lookupNativeSession(provider, sessionID)
}

func (a *FSAdapter) lookupNativeSession(provider, sessionID string) (NativeSession, bool, error) {
	root, err := os.OpenRoot(a.root)
	if errors.Is(err, os.ErrNotExist) {
		return NativeSession{}, false, nil
	}
	if err != nil {
		return NativeSession{}, false, err
	}
	defer root.Close()
	file, err := root.Open(nativeSessionPath(provider, sessionID))
	if errors.Is(err, os.ErrNotExist) {
		return NativeSession{}, false, nil
	}
	if err != nil {
		return NativeSession{}, false, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 16384 {
		return NativeSession{}, false, errors.New("workdir: native session binding must be a regular file under 16 KiB")
	}
	var record NativeSession
	decoder := json.NewDecoder(io.LimitReader(file, 16385))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&record); err != nil {
		return record, false, errors.New("workdir: invalid native session binding")
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return record, false, errors.New("workdir: trailing native session binding data")
	}
	if err := validateNativeSession(record); err != nil || record.Provider != provider || record.SessionID != sessionID {
		return record, false, errors.New("workdir: native session binding identity mismatch")
	}
	record.Workdir = workspaceForRun(a.root, TaskID{Workspace: record.WorkspaceID, Task: record.TaskID}, record.OriginRunID).Workdir
	return record, true, nil
}

func validateNativeSession(record NativeSession) error {
	_, err := prepareRunID(TaskID{Workspace: record.WorkspaceID, Task: record.TaskID, Run: record.OriginRunID})
	if err != nil || record.Schema != 1 || strings.TrimSpace(record.Provider) == "" ||
		strings.TrimSpace(record.SessionID) == "" || record.OriginRunID == "" || record.GrantHash == "" || record.Fingerprint == "" {
		return errors.New("workdir: incomplete native session binding")
	}
	return nil
}

func (a *FSAdapter) PinNativeSession(record NativeSession) error {
	if err := validateNativeSession(record); err != nil {
		return err
	}
	nativeSessionMu.Lock()
	defer nativeSessionMu.Unlock()
	previous, found, err := a.lookupNativeSession(record.Provider, record.SessionID)
	if err != nil {
		return err
	}
	previous.Workdir, record.Workdir = "", ""
	if found {
		if previous != record {
			return errors.New("workdir: native session already belongs to another execution context")
		}
		return nil
	}
	if err := os.MkdirAll(a.root, 0o755); err != nil {
		return err
	}
	root, err := os.OpenRoot(a.root)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.Mkdir(".native-sessions", 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if len(data) > 16384 {
		return errors.New("workdir: native session binding exceeds 16 KiB")
	}
	path := nativeSessionPath(record.Provider, record.SessionID)
	tmp := path + "." + rand.Text()
	file, err := root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = root.Remove(tmp) }()
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return fmt.Errorf("workdir: write native session binding: %w", err)
	}
	return root.Rename(tmp, path)
}
