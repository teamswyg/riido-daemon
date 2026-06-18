package taskdb

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/teamswyg/riido-daemon/pkg/util/fileutil"
)

func DefaultTaskDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "riido", "task-db.json"), nil
}

func LoadTaskDBOrEmpty(path string) (TaskDB, error) {
	db, err := LoadTaskDB(path)
	if os.IsNotExist(err) {
		return EmptyTaskDB(), nil
	}
	return db, err
}

func SaveTaskDB(path string, db TaskDB) error {
	if path == "" {
		return taskDBErrorf(ErrTaskDBInput, "save", "task DB path is empty")
	}
	if err := fileutil.WriteJSONAtomic(path, normalizeTaskDB(db)); err != nil {
		return taskDBWrapf(ErrTaskDBPersistence, "save", err, "save task DB")
	}
	return nil
}

func LoadTaskDB(path string) (TaskDB, error) {
	var db TaskDB
	data, err := os.ReadFile(path)
	if err != nil {
		return db, err
	}
	if err := json.Unmarshal(data, &db); err != nil {
		return db, taskDBWrapf(ErrTaskDBPersistence, "load.decode", err, "decode task DB")
	}
	if db.SchemaVersion != TaskDBSchemaVersion {
		return db, taskDBErrorf(ErrTaskDBSchema, "load.validate-schema", "task DB schema mismatch: got %q want %q", db.SchemaVersion, TaskDBSchemaVersion)
	}
	return normalizeTaskDB(db), nil
}
