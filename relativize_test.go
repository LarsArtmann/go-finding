package finding_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
)

// The SDK lane must produce the same finding ID the CLI lane produces for the
// same source fact: absolute paths under the root relativize, and the
// regenerated ID matches the one built from the relative path directly
// (the checkout-dependence documented in the GenerateID note).
func TestRelativizeSDKLaneMatchesCLILane(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	absFile := filepath.Join(root, "internal", "config", "config.go")

	sdkFinding := finding.Finding{
		ToolName: "branching-flow",
		Rule:     "DUPLICATE_TYPE",
		Position: finding.Position{File: finding.FilePath(absFile), Line: 179},
	}
	sdkFinding.ID = finding.GenerateID(sdkFinding.ToolName, sdkFinding.Rule, sdkFinding.Position)

	cliFinding := finding.Finding{
		ToolName: "branching-flow",
		Rule:     "DUPLICATE_TYPE",
		Position: finding.Position{File: finding.FilePath("internal/config/config.go"), Line: 179},
	}
	cliFinding.ID = finding.GenerateID(cliFinding.ToolName, cliFinding.Rule, cliFinding.Position)

	got := finding.Relativize([]finding.Finding{sdkFinding}, root)

	if got[0].ID != cliFinding.ID {
		t.Errorf("SDK-lane ID %q must equal CLI-lane ID %q after relativization", got[0].ID, cliFinding.ID)
	}

	if string(got[0].Position.File) != "internal/config/config.go" {
		t.Errorf("position must be root-relative, got %q", got[0].Position.File)
	}
}

func TestRelativizeLeavesUnrelatedPathsAlone(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	other := t.TempDir()

	outside := finding.Finding{
		ToolName: "branching-flow",
		Rule:     "PHANTOM-TRANSPOSE",
		Position: finding.Position{File: finding.FilePath(filepath.Join(other, "x.go")), Line: 3},
	}
	outside.ID = finding.GenerateID(outside.ToolName, outside.Rule, outside.Position)

	relative := finding.Finding{
		ToolName: "branching-flow",
		Rule:     "PHANTOM-TRANSPOSE",
		Position: finding.Position{File: finding.FilePath("pkg/x.go"), Line: 3},
	}
	relative.ID = finding.GenerateID(relative.ToolName, relative.Rule, relative.Position)

	wantOutsideID, wantRelativeID := outside.ID, relative.ID

	got := finding.Relativize([]finding.Finding{outside, relative}, root)

	if got[0].ID != wantOutsideID || string(got[0].Position.File) == "x.go" {
		t.Errorf("absolute path outside root must stay untouched, got ID %q file %q",
			got[0].ID, got[0].Position.File)
	}

	if got[1].ID != wantRelativeID {
		t.Errorf("already-relative path must keep its ID, got %q want %q", got[1].ID, wantRelativeID)
	}
}

// A relative root (the CLI's `scan .` form) must still relativize absolute
// finding paths, and an empty root must be a no-op.
func TestRelativizeRelativeRootAndEmptyRootNoOp(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	f := finding.Finding{
		ToolName: "branching-flow",
		Rule:     "COMPOSITION_mixin",
		Position: finding.Position{File: finding.FilePath(filepath.Join(wd, "filter.go")), Line: 1, Column: 2},
	}
	f.ID = finding.GenerateID(f.ToolName, f.Rule, f.Position)

	got := finding.Relativize([]finding.Finding{f}, ".")
	if string(got[0].Position.File) != "filter.go" {
		t.Errorf("relative root must relativize, got %q", got[0].Position.File)
	}

	untouched := finding.Relativize([]finding.Finding{f}, "")
	if untouched[0].ID != f.ID {
		t.Errorf("empty root must be a no-op, got ID %q want %q", untouched[0].ID, f.ID)
	}
}

// Relativize is pure: the input slice must survive the call byte-for-byte so
// callers can keep using their original findings (e.g. for an absolute-path
// report) alongside the relativized copy.
func TestRelativizeDoesNotMutateInput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	original := finding.Finding{
		ToolName: "branching-flow",
		Rule:     "FLAG_PARAM",
		Position: finding.Position{File: finding.FilePath(filepath.Join(root, "a", "b.go")), Line: 7},
	}
	original.ID = finding.GenerateID(original.ToolName, original.Rule, original.Position)

	input := []finding.Finding{original}
	wantFile, wantID := string(original.Position.File), original.ID

	_ = finding.Relativize(input, root)

	if string(input[0].Position.File) != wantFile || input[0].ID != wantID {
		t.Errorf("input mutated: file %q (want %q), ID %q (want %q)",
			input[0].Position.File, wantFile, input[0].ID, wantID)
	}

	if n := finding.Relativize(nil, root); n != nil {
		t.Errorf("nil input must yield nil, got %v", n)
	}
}

// Line-0 file-level findings carry hash IDs over (tool, rule, file):
// relativizing the file regenerates the hash (the old one embedded the host
// path), and the regenerated hash is identical across two different checkouts
// of the same layout.
func TestRelativizeRegeneratesHashIDsForFileLevelFindings(t *testing.T) {
	t.Parallel()

	rootOne, rootTwo := t.TempDir(), t.TempDir()

	f := func(root string) finding.Finding {
		x := finding.Finding{
			ToolName: "branching-flow",
			Rule:     "PROJECT_CHECK",
			Position: finding.Position{File: finding.FilePath(filepath.Join(root, "pkg", "cfg.go"))},
		}
		x.ID = finding.GenerateID(x.ToolName, x.Rule, x.Position)
		return x
	}

	one, two := f(rootOne), f(rootTwo)

	if one.ID == two.ID {
		t.Fatalf("precondition: host-path hash IDs must differ across roots, both %q", one.ID)
	}

	gotOne := finding.Relativize([]finding.Finding{one}, rootOne)[0]
	gotTwo := finding.Relativize([]finding.Finding{two}, rootTwo)[0]

	if gotOne.ID != gotTwo.ID {
		t.Errorf("relativized hash IDs must agree across hosts: %q vs %q", gotOne.ID, gotTwo.ID)
	}

	if !finding.IsHashID(string(gotOne.ID)) {
		t.Errorf("relativized file-level ID must stay hash-based, got %q", gotOne.ID)
	}
}
