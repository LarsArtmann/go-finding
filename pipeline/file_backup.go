package pipeline

import (
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/larsartmann/go-finding"
)

// FileBackup manages backup and restore of files during fix application.
// It is safe for concurrent use.
type FileBackup struct {
	enabled   bool
	backupDir string
	backups   map[string]string // original -> backup path
	mu        sync.Mutex
}

// NewFileBackup creates a FileBackup that stores backups in the given directory.
func NewFileBackup(backupDir string) *FileBackup {
	return &FileBackup{
		enabled:   true,
		backupDir: backupDir,
		backups:   make(map[string]string),
	}
}

// IsEnabled reports whether backup is enabled.
func (fb *FileBackup) IsEnabled() bool {
	return fb.enabled
}

// SetEnabled controls whether backups are created.
func (fb *FileBackup) SetEnabled(v bool) {
	fb.enabled = v
}

// BackupPath returns the backup path for the given original file, or empty
// string if no backup exists.
func (fb *FileBackup) BackupPath(original string) string {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	return fb.backups[original]
}

// Backup creates a backup of the given file.
func (fb *FileBackup) Backup(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return ioErrorAt("read file for backup", err, path)
	}

	backupPath := filepath.Join(
		fb.backupDir,
		fmt.Sprintf("%x_%d.bak", fileHash(path), time.Now().UnixNano()),
	)
	if err := os.MkdirAll(fb.backupDir, 0o750); err != nil {
		return finding.NewIOError("create backup dir", err)
	}

	if err := os.WriteFile( //nolint:gosec // intentional file write in fix applier
		backupPath,
		data,
		0o600,
	); err != nil {
		return ioErrorAt("write backup", err, path)
	}

	fb.mu.Lock()
	fb.backups[path] = backupPath
	fb.mu.Unlock()

	return nil
}

// Restore restores a file from its backup.
func (fb *FileBackup) Restore(path string) error {
	backupPath, ok := func() (string, bool) {
		fb.mu.Lock()
		defer fb.mu.Unlock()

		p, exists := fb.backups[path]

		return p, exists
	}()
	if !ok {
		return finding.NewInternalError("no backup for "+path, nil)
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return ioErrorAt("read backup", err, path)
	}

	if err := os.WriteFile( //nolint:gosec // intentional file write in fix applier
		path,
		data,
		0o600,
	); err != nil {
		return ioErrorAt("restore file", err, path)
	}

	return nil
}

// RollbackAll restores all modified files from their backups.
func (fb *FileBackup) RollbackAll(paths []string) error {
	var errs []error

	for _, p := range paths {
		if err := fb.Restore(p); err != nil {
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
