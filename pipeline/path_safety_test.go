package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
)

func TestResolveSafePath_NormalRelativePath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	rel := "src/main.go"

	got, ok := resolveSafePath(root, rel)
	if !ok {
		t.Fatalf("expected safe, got unsafe")
	}

	want := filepath.Clean(filepath.Join(root, rel))
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveSafePath_PathTraversal(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	cases := []struct {
		name string
		rel  string
	}{
		{"escape parent", "../../../etc/passwd"},
		{"escape single", "../secret"},
		{"nested escape", "a/b/../../../etc/shadow"},
		{"double dot prefix", "../../.ssh/id_rsa"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, ok := resolveSafePath(root, tc.rel)
			if ok {
				t.Errorf("rel=%q: expected unsafe, got safe", tc.rel)
			}
		})
	}
}

func TestResolveSafePath_AbsolutePathTreatedAsRelative(t *testing.T) {
	// Absolute paths are normalized to relative by filepath.Join,
	// so they cannot escape rootDir. This is a deliberate security property.
	t.Parallel()

	root := t.TempDir()

	cases := []struct {
		name string
		abs  string
	}{
		{"/etc/passwd", "/etc/passwd"},
		{"absolute inside root", filepath.Join(root, "file.go")},
		{"/tmp/secret", "/tmp/secret"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := resolveSafePath(root, tc.abs)
			if !ok {
				t.Errorf("abs=%q: expected safe (normalized under root), got unsafe", tc.abs)
			}

			// The result should be under root, not the raw absolute path.
			if got == tc.abs {
				t.Errorf("abs=%q: result should be normalized under root, got raw abs path", tc.abs)
			}
		})
	}
}

func TestResolveSafePath_EmptyRelPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	got, ok := resolveSafePath(root, "")
	if !ok {
		t.Fatalf("empty rel should resolve to root, expected safe")
	}

	cleanRoot, _ := filepath.EvalSymlinks(filepath.Clean(root))
	if got != cleanRoot {
		t.Errorf("got %q, want %q", got, cleanRoot)
	}
}

func TestResolveSafePath_DotPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	got, ok := resolveSafePath(root, ".")
	if !ok {
		t.Fatalf("'.' should resolve to root, expected safe")
	}

	cleanRoot, _ := filepath.EvalSymlinks(filepath.Clean(root))
	if got != cleanRoot {
		t.Errorf("got %q, want %q", got, cleanRoot)
	}
}

func TestResolveSafePath_SymlinkInsideRoot_PointingInside(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	realFile := filepath.Join(root, "real.go")

	err := os.WriteFile(realFile, []byte("x"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(root, "link.go")

	err = os.Symlink(realFile, link)
	if err != nil {
		t.Fatal(err)
	}

	got, ok := resolveSafePath(root, "link.go")
	if !ok {
		t.Fatalf("symlink inside root pointing inside should be safe")
	}

	resolved, _ := filepath.EvalSymlinks(link)
	if got != resolved {
		t.Errorf("got %q, want resolved %q", got, resolved)
	}
}

func TestResolveSafePath_SymlinkInsideRoot_PointingOutside(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	outsideDir := t.TempDir()

	outsideFile := filepath.Join(outsideDir, "secret")

	err := os.WriteFile(outsideFile, []byte("secret"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(root, "escape.go")

	err = os.Symlink(outsideFile, link)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := resolveSafePath(root, "escape.go")
	if ok {
		t.Errorf("symlink pointing outside root should be unsafe")
	}
}

func TestResolveSafePath_NonexistentFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	got, ok := resolveSafePath(root, "nonexistent.go")
	if !ok {
		t.Fatalf("nonexistent file inside root should be safe (path is still within root)")
	}

	want := filepath.Clean(filepath.Join(root, "nonexistent.go"))
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveSafePath_RootItself(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	got, ok := resolveSafePath(root, "")
	if !ok {
		t.Fatalf("root itself should be safe")
	}

	cleanRoot, _ := filepath.EvalSymlinks(filepath.Clean(root))
	if got != cleanRoot {
		t.Errorf("got %q, want %q", got, cleanRoot)
	}
}

func TestResolveSafePath_DeeplyNestedPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	rel := "a/b/c/d/e/f.go"

	got, ok := resolveSafePath(root, rel)
	if !ok {
		t.Fatalf("deeply nested path inside root should be safe")
	}

	want := filepath.Clean(filepath.Join(root, rel))
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestGroupFindingsBySafePath_ResolvedPathAsMapKey verifies the TOCTOU
// mitigation: the map key is the resolved real path (after EvalSymlinks),
// not the raw symlink path. This ensures file I/O uses the path captured
// at validation time, so a symlink swap between validation and I/O cannot
// redirect writes outside rootDir.
func TestGroupFindingsBySafePath_ResolvedPathAsMapKey(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	realFile := filepath.Join(root, "real.go")
	if err := os.WriteFile(realFile, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	linkPath := filepath.Join(root, "link.go")
	if err := os.Symlink(realFile, linkPath); err != nil {
		t.Fatal(err)
	}

	applier, err := NewFixApplier(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = applier.Close() }()

	fixes := []finding.Finding{
		finding.NewFinding(
			finding.RuleName("rule1"),
			finding.ToolName("test"),
			"test",
			finding.SeverityWarning,
			finding.Pos(finding.FilePath("link.go"), 1, 1),
			finding.Confidence(0.9),
		),
	}

	byFile := applier.groupFindingsBySafePath(fixes)

	resolved, _ := filepath.EvalSymlinks(linkPath)

	for key := range byFile {
		if key == linkPath {
			t.Errorf("map key should be resolved real path %q, got raw symlink path %q",
				resolved, key)
		}

		if key != resolved {
			t.Errorf("map key = %q, want resolved path %q", key, resolved)
		}
	}

	if len(byFile) != 1 {
		t.Fatalf("expected 1 file group, got %d", len(byFile))
	}
}

// TestGroupFindingsBySafePath_PathTraversalFiltered verifies that findings
// with path traversal in Position.File are silently skipped.
func TestGroupFindingsBySafePath_PathTraversalFiltered(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, "safe.go"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	applier, err := NewFixApplier(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = applier.Close() }()

	fixes := []finding.Finding{
		finding.NewFinding(
			finding.RuleName("rule-safe"),
			finding.ToolName("test"),
			"safe",
			finding.SeverityWarning,
			finding.Pos(finding.FilePath("safe.go"), 1, 1),
			finding.Confidence(0.9),
		),
		finding.NewFinding(
			finding.RuleName("rule-escape"),
			finding.ToolName("test"),
			"escape",
			finding.SeverityWarning,
			finding.Pos(finding.FilePath("../../../etc/passwd"), 1, 1),
			finding.Confidence(0.9),
		),
	}

	byFile := applier.groupFindingsBySafePath(fixes)

	if len(byFile) != 1 {
		t.Fatalf("expected 1 safe file (traversal filtered), got %d groups", len(byFile))
	}

	for _, findings := range byFile {
		if len(findings) != 1 {
			t.Errorf("expected 1 finding in safe group, got %d", len(findings))
		}

		if findings[0].Rule != "rule-safe" {
			t.Errorf("expected rule-safe, got %s", findings[0].Rule)
		}
	}
}
