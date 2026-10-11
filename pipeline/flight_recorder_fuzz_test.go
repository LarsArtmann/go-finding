package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// FuzzSanitizeFilename verifies that sanitizeFilename always produces a safe,
// non-empty filename consisting only of alphanumeric characters and single hyphens.
func FuzzSanitizeFilename(f *testing.F) {
	f.Add("detect")
	f.Add("slow-stage-30s")
	f.Add("")
	f.Add("!!!")
	f.Add("a---b")
	f.Add("/etc/passwd")
	f.Add("café-münchen")
	f.Add("stage\u0000null")
	f.Add("very-long-reason-" + strings.Repeat("x", 200))

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 500 {
			return
		}

		result := sanitizeFilename(input)

		if result == "" {
			t.Fatal("sanitizeFilename returned empty string")
		}

		if strings.Contains(result, "--") {
			t.Fatalf("result contains consecutive hyphens: %q", result)
		}

		if strings.HasPrefix(result, "-") || strings.HasSuffix(result, "-") {
			t.Fatalf("result has leading/trailing hyphen: %q", result)
		}

		for _, r := range result {
			isAlphaNum := (r >= 'a' && r <= 'z') ||
				(r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') ||
				r == '-'
			if !isAlphaNum {
				t.Fatalf("result contains non-safe character %q in %q", r, result)
			}
		}
	})
}

// FuzzPruneSnapshotsHostileDir verifies pruneSnapshots against hostile output
// directory contents: fuzzed filename components, directories masquerading as
// snapshots, dangling symlinks, and symlinks pointing outside the output dir.
// Invariants: no panic, non-matching files and symlink targets are never
// touched, and the own-snapshot count never exceeds MaxFiles.
func FuzzPruneSnapshotsHostileDir(f *testing.F) {
	f.Add([]byte("hostile"))
	f.Add([]byte("a/b\\c:d*e"))
	f.Add([]byte(".."))
	f.Add([]byte("\x00\x01\xff"))
	f.Add([]byte(strings.Repeat("x", 300)))

	f.Fuzz(func(t *testing.T, seed []byte) {
		name := strings.Map(func(r rune) rune {
			if r == 0 || r == '/' || r == '\\' {
				return 'x'
			}
			return r
		}, string(seed))
		if len(name) > 200 {
			name = name[:200]
		}
		if name == "" {
			name = "empty"
		}

		dir := t.TempDir()

		unrelated := filepath.Join(dir, "unrelated.txt")
		foreign := filepath.Join(dir, "other-tool-000.trace")
		dirTrap := filepath.Join(dir, "go-finding-trace-dir.trace")
		for _, p := range []string{unrelated, foreign} {
			if err := os.WriteFile(p, []byte("p"), 0o600); err != nil {
				t.Fatalf("write protected file: %v", err)
			}
		}
		if err := os.Mkdir(dirTrap, 0o755); err != nil {
			t.Fatalf("mkdir trap: %v", err)
		}

		target := filepath.Join(t.TempDir(), "target.txt")
		if err := os.WriteFile(target, []byte("p"), 0o600); err != nil {
			t.Fatalf("write target: %v", err)
		}
		if err := os.Symlink(target, filepath.Join(dir, "go-finding-trace-link.trace")); err != nil {
			t.Fatalf("symlink: %v", err)
		}
		if err := os.Symlink(
			filepath.Join(dir, "missing"),
			filepath.Join(dir, "go-finding-trace-dangling.trace"),
		); err != nil {
			t.Fatalf("dangling symlink: %v", err)
		}

		own := filepath.Join(dir, "go-finding-trace-000-"+name+".trace")
		if err := os.WriteFile(own, []byte("old"), 0o600); err != nil {
			t.Fatalf("write own file: %v", err)
		}
		past := time.Now().Add(-time.Hour)
		if err := os.Chtimes(own, past, past); err != nil {
			t.Fatalf("chtimes: %v", err)
		}

		hook, err := NewFlightRecorderHook(FlightRecorderConfig{
			OutputDir: dir,
			MaxFiles:  1,
		})
		if err != nil {
			t.Fatalf("NewFlightRecorderHook: %v", err)
		}

		if _, err := hook.Snapshot(context.Background(), "prune"); err != nil {
			t.Fatalf("Snapshot: %v", err)
		}
		hook.Close()

		for _, p := range []string{unrelated, foreign, dirTrap, target} {
			if _, err := os.Lstat(p); err != nil {
				t.Fatalf("protected path %s affected by pruning: %v", p, err)
			}
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("ReadDir: %v", err)
		}
		matching := 0
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			n := e.Name()
			if strings.HasPrefix(n, "go-finding-trace-") &&
				(strings.HasSuffix(n, ".trace") || strings.HasSuffix(n, ".trace.gz")) {
				matching++
			}
		}
		if matching > 1 {
			t.Fatalf("expected at most %d own snapshot files after prune, got %d", 1, matching)
		}
	})
}
