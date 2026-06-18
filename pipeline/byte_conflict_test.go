package pipeline

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/larsartmann/go-finding"
)

func TestByteLevelConflictDetection(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		fix1Before      string
		fix1After       string
		fix1Line        int
		fix2Before      string
		fix2After       string
		fix2Line        int
		wantConflicts   int
		wantApplied     int
		expectF1Applied bool
	}{
		{
			name:       "conflicting edits on same byte range",
			fix1Before: "hello", fix1After: "world", fix1Line: 4,
			fix2Before: "hello", fix2After: "universe", fix2Line: 4,
			wantConflicts:   1,
			wantApplied:     1,
			expectF1Applied: true,
		},
		{
			name:       "non-conflicting edits on different byte ranges",
			fix1Before: "hello", fix1After: "world", fix1Line: 4,
			fix2Before: "main", fix2After: "myApp", fix2Line: 3,
			wantConflicts: 0,
			wantApplied:   2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := "fixme.go"
			writeTestFile(t, filepath.Join(tmpDir, testFile),
				[]byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n"))

			fix1 := makeFixFinding("f1", tc.fix1Before, tc.fix1After, testFile, tc.fix1Line)
			fix2 := makeFixFinding("f2", tc.fix2Before, tc.fix2After, testFile, tc.fix2Line)

			cfg := Config{
				MaxIterations:              1,
				ByteLevelConflictDetection: true,
			}

			if tc.expectF1Applied {
				cfg.OnFix = func(f finding.Finding, applied bool) {
					if f.ID == "f1" && !applied {
						t.Errorf("f1 should have been applied")
					}
				}
			}

			det := mockDetWithFindings("tool", fix1, fix2)

			p, err := New(cfg, tmpDir, det)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			result, err := p.Run(context.Background())
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			if len(result.Iterations) == 0 {
				t.Fatal("expected at least one iteration")
			}

			if result.Iterations[0].Conflicts != tc.wantConflicts {
				t.Errorf("expected %d conflicts, got %d", tc.wantConflicts, result.Iterations[0].Conflicts)
			}

			if result.Iterations[0].Applied != tc.wantApplied {
				t.Errorf("expected %d applied, got %d", tc.wantApplied, result.Iterations[0].Applied)
			}
		})
	}
}

func TestByteLevelConflictDetection_MissingFile_FailOpen(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	fix1 := makeFixFinding("f1", "hello", "world", "nonexistent.go", 4)

	cfg := Config{
		MaxIterations:              1,
		ByteLevelConflictDetection: true,
	}

	det := mockDetWithFindings("tool", fix1)

	p, err := New(cfg, tmpDir, det)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = p.Run(context.Background())
	if err == nil {
		t.Fatal("expected error when applying fix to nonexistent file")
	}
}
