package pipeline

import (
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/lockutil"
)

type backupEntry struct {
	path string
	mode os.FileMode
}

// FileBackup manages backup and restore of files during fix application.
// It is safe for concurrent use.
type FileBackup struct {
	enabled   atomic.Bool
	backupDir string
	backups   map[string]backupEntry // original path -> backup entry
	mu        sync.Mutex
}

// NewFileBackup creates a FileBackup that stores backups in the given directory.
func NewFileBackup(backupDir string) *FileBackup {
	fb := &FileBackup{ //nolint:exhaustruct
		backupDir: backupDir,
		backups:   make(map[string]backupEntry),
	}
	fb.enabled.Store(true)

	return fb
}

// IsEnabled reports whether backup is enabled.
func (fb *FileBackup) IsEnabled() bool {
	return fb.enabled.Load()
}

// SetEnabled controls whether backups are created.
func (fb *FileBackup) SetEnabled(v bool) {
	fb.enabled.Store(v)
}

// BackupPath returns the backup path for the given original file, or empty
// string if no backup exists.
func (fb *FileBackup) BackupPath(original string) string {
	return lockutil.Locked(&fb.mu, func() string {
		if e, ok := fb.backups[original]; ok {
			return e.path
		}

		return ""
	})
}

// Backup creates a backup of the given file.
func (fb *FileBackup) Backup(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return ioErrorAt("open file for backup", err, path)
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(f)
	if err != nil {
		return ioErrorAt("read file for backup", err, path)
	}

	info, err := f.Stat()
	if err != nil {
		return ioErrorAt("stat file for backup", err, path)
	}

	backupPath := filepath.Join(
		fb.backupDir,
		fmt.Sprintf("%x_%d.bak", fileHash(path), time.Now().UnixNano()),
	)

	err = os.MkdirAll(fb.backupDir, 0o750)
	if err != nil {
		return finding.NewIOError("create backup dir", err)
	}

	err = os.WriteFile(
		backupPath,
		data,
		0o600,
	)
	if err != nil {
		return ioErrorAt("write backup", err, path)
	}

	lockutil.Locked(&fb.mu, func() struct{} {
		fb.backups[path] = backupEntry{path: backupPath, mode: info.Mode()}

		return struct{}{}
	})

	return nil
}

// Restore restores a file from its backup.
func (fb *FileBackup) Restore(path string) error {
	type lookup struct {
		entry backupEntry
		ok    bool
	}

	res := lockutil.Locked(&fb.mu, func() lookup {
		e, exists := fb.backups[path]

		return lookup{entry: e, ok: exists}
	})
	if !res.ok {
		return finding.NewInternalError("no backup for "+path, nil)
	}

	data, err := os.ReadFile(res.entry.path)
	if err != nil {
		return ioErrorAt("read backup", err, path)
	}

	err = os.WriteFile( //nolint:gosec // intentional file write in fix applier
		path,
		data,
		res.entry.mode,
	)
	if err != nil {
		return ioErrorAt("restore file", err, path)
	}

	return nil
}

// RollbackAll restores all modified files from their backups.
func (fb *FileBackup) RollbackAll(paths []string) error {
	var errs []error

	for _, p := range paths {
		err := fb.Restore(p)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// fileHash returns the hex-encoded FNV-1a 128-bit hash of s.
func fileHash(s string) string {
	h := fnv.New128a()
	h.Write([]byte(s))

	return hex.EncodeToString(h.Sum(nil))
}
