package toolsdk

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
)

func TestRegisterAndAll(t *testing.T) {
	ResetForTest()

	detector := finding.NamedDetectorFunc("test-tool", func(_ context.Context) ([]finding.Finding, error) {
		return nil, nil
	})

	Register(Spec{
		Name:        "test-tool",
		Description: "a test tool",
		Trigger:     OnGoFiles(),
		Inputs:      []string{"**/*.go"},
		Detect:      detector,
	})

	all := All()
	if len(all) != 1 {
		t.Fatalf("expected 1 registered spec, got %d", len(all))
	}

	if all[0].Name != "test-tool" {
		t.Errorf("expected test-tool, got %s", all[0].Name)
	}

	if all[0].Trigger.Language != "go" {
		t.Errorf("expected go language, got %s", all[0].Trigger.Language)
	}
}

// TestRegisterAllRoundTripFullTriggerFidelity pins the registry round trip:
// EVERY Trigger field a spec registers must come back from All() unchanged.
// The earlier test asserted Language only, so a registry copy that dropped
// NotRequires (the ownership-deference field) would have gone unnoticed.
func TestRegisterAllRoundTripFullTriggerFidelity(t *testing.T) {
	snapshot := SnapshotForTest()
	t.Cleanup(func() { RestoreForTest(snapshot) })
	ResetForTest()

	detector := finding.NamedDetectorFunc("trigger-fidelity-tool", func(context.Context) ([]finding.Finding, error) {
		return nil, nil
	})

	want := Trigger{
		Files:       []string{"**/*.go", "go.mod"},
		Language:    "go",
		Requires:    []string{"**/go.mod", "**/go.work"},
		NotRequires: []string{"**/vendor/**", "**/testdata/**"},
	}

	Register(Spec{
		Name:        "trigger-fidelity-tool",
		Description: "round-trip fidelity probe",
		Trigger:     want,
		Inputs:      []string{"**/*.go"},
		Outputs:     []string{"**/*.go"},
		Detect:      detector,
	})

	all := All()
	if len(all) != 1 {
		t.Fatalf("expected 1 registered spec, got %d", len(all))
	}

	got := all[0].Trigger

	if got.Language != want.Language {
		t.Errorf("Language round trip: got %q, want %q", got.Language, want.Language)
	}

	if !slices.Equal(got.Files, want.Files) {
		t.Errorf("Files round trip: got %v, want %v", got.Files, want.Files)
	}

	if !slices.Equal(got.Requires, want.Requires) {
		t.Errorf("Requires round trip: got %v, want %v", got.Requires, want.Requires)
	}

	if !slices.Equal(got.NotRequires, want.NotRequires) {
		t.Errorf("NotRequires round trip: got %v, want %v", got.NotRequires, want.NotRequires)
	}

	if !slices.Equal(all[0].Outputs, []string{"**/*.go"}) {
		t.Errorf("Outputs round trip: got %v, want %v", all[0].Outputs, []string{"**/*.go"})
	}
}

func TestRegisterRequiresName(t *testing.T) {
	ResetForTest()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on empty Name")
		}
	}()

	Register(Spec{
		Detect: finding.NamedDetectorFunc("x", func(context.Context) ([]finding.Finding, error) { return nil, nil }),
	})
}

func TestRegisterRequiresCapability(t *testing.T) {
	ResetForTest()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on no Detect/Repair")
		}
	}()

	Register(Spec{Name: "empty", Description: "has description"})
}

func TestRegisterRequiresDescription(t *testing.T) {
	ResetForTest()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on empty Description")
		}
	}()

	Register(Spec{
		Name:   "no-desc",
		Detect: finding.NamedDetectorFunc("x", func(context.Context) ([]finding.Finding, error) { return nil, nil }),
	})
}

func TestRegisterRequiresDescriptionRejectsWhitespace(t *testing.T) {
	ResetForTest()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on whitespace-only Description")
		}
	}()

	Register(Spec{
		Name:        "ws-desc",
		Description: "   ",
		Detect: finding.NamedDetectorFunc("x", func(context.Context) ([]finding.Finding, error) {
			return nil, nil
		}),
	})
}

func TestAllReturnsCopy(t *testing.T) {
	ResetForTest()

	detector := finding.NamedDetectorFunc("t1", func(context.Context) ([]finding.Finding, error) { return nil, nil })
	Register(Spec{Name: "t1", Description: "test tool", Detect: detector})

	all := All()
	all[0].Name = "mutated"

	if All()[0].Name != "t1" {
		t.Error("All() should return a copy immune to caller mutation")
	}
}

func TestOnGoFiles(t *testing.T) {
	tr := OnGoFiles()
	if tr.Language != "go" || len(tr.Files) != 1 || tr.Files[0] != "**/*.go" {
		t.Errorf("unexpected OnGoFiles trigger: %+v", tr)
	}
}

// TestSpecCarriesModuleFanOut pins the Spec field surface that mirrors
// BuildFlow's domain/tool.DAGTopology: ModuleFanOut declares per-module
// execution in multi-module workspaces, and consumers convert it
// field-for-field at spec-to-tool conversion time.
func TestSpecCarriesModuleFanOut(t *testing.T) {
	ResetForTest()

	detector := finding.NamedDetectorFunc("fanout-tool", func(context.Context) ([]finding.Finding, error) {
		return nil, nil
	})

	Register(Spec{
		Name:         "fanout-tool",
		Description:  "runs per Go module",
		Trigger:      OnGoFiles(),
		DependsOn:    []string{"go-fix"},
		ModuleFanOut: true,
		Detect:       detector,
	})

	got := All()[0]
	if !got.ModuleFanOut {
		t.Error("ModuleFanOut must round-trip through the registry")
	}

	plain := Spec{Name: "t", Description: "d", Detect: detector}
	if plain.ModuleFanOut {
		t.Error("zero-value Spec must default ModuleFanOut to false")
	}
}

// TestTriggerCarriesNotRequires pins the Trigger field surface that mirrors
// BuildFlow's domain/tool.Trigger: NotRequires expresses ownership
// deference (disqualifying patterns), and a spec that sets it must be able
// to round-trip the value through the struct — consumers convert it
// field-for-field.
func TestTriggerCarriesNotRequires(t *testing.T) {
	want := []string{"**/treefmt.toml", "**/.treefmt.toml"}

	got := Trigger{
		Files:       []string{"**/*.nix"},
		Language:    "nix",
		Requires:    []string{"flake.nix"},
		NotRequires: want,
	}

	if len(got.NotRequires) != len(want) {
		t.Fatalf("NotRequires length = %d, want %d", len(got.NotRequires), len(want))
	}

	for i := range want {
		if got.NotRequires[i] != want[i] {
			t.Errorf("NotRequires[%d] = %q, want %q", i, got.NotRequires[i], want[i])
		}
	}
}

// TestSpecCarriesSwitchCases pins the R6 surface: SwitchCases mirrors
// BuildFlow's domain/tool.SwitchCase (Suffix + FindingRuleIDs +
// ExcludedRuleIDs) so SDK specs can express gotcha-#194 mutual-exclusion
// topologies instead of racing mutating repair cases.
func TestSpecCarriesSwitchCases(t *testing.T) {
	want := SwitchCase{
		Suffix:          "fix-hash-mismatch",
		FindingRuleIDs:  []string{"fod-hash-mismatch"},
		ExcludedRuleIDs: []string{"gomod-stale", "vendor-inconsistency"},
	}

	got := Spec{
		Name:         "example",
		Description:  "example tool",
		ModuleFanOut: false,
		SwitchCases:  []SwitchCase{want},
	}

	if len(got.SwitchCases) != 1 {
		t.Fatalf("SwitchCases length = %d, want 1", len(got.SwitchCases))
	}

	c := got.SwitchCases[0]

	if c.Suffix != want.Suffix {
		t.Errorf("Suffix = %q, want %q", c.Suffix, want.Suffix)
	}

	if len(c.FindingRuleIDs) != len(want.FindingRuleIDs) || c.FindingRuleIDs[0] != want.FindingRuleIDs[0] {
		t.Errorf("FindingRuleIDs = %v, want %v", c.FindingRuleIDs, want.FindingRuleIDs)
	}

	if len(c.ExcludedRuleIDs) != len(want.ExcludedRuleIDs) || c.ExcludedRuleIDs[0] != want.ExcludedRuleIDs[0] {
		t.Errorf("ExcludedRuleIDs = %v, want %v", c.ExcludedRuleIDs, want.ExcludedRuleIDs)
	}
}

// TestTriggerValidate covers Trigger.Validate (issue #44): Language must be
// "" or one of ValidLanguages, and Files/Requires/NotRequires must not
// contain empty patterns. Includes the natural "golang" typo that used to
// register cleanly and then silently never match downstream.
func TestTriggerValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		trigger  Trigger
		wantErr  error
		contains string
	}{
		{name: "zero value is language-agnostic", trigger: Trigger{}},
		{name: "language go", trigger: Trigger{Language: "go"}},
		{name: "language js", trigger: Trigger{Language: "js"}},
		{name: "language ts", trigger: Trigger{Language: "ts"}},
		{name: "language python", trigger: Trigger{Language: "python"}},
		{name: "language nix", trigger: Trigger{Language: "nix"}},
		{name: "language rust", trigger: Trigger{Language: "rust"}},
		{
			name:     "golang typo",
			trigger:  Trigger{Language: "golang"},
			wantErr:  ErrInvalidTrigger,
			contains: `Language "golang"`,
		},
		{
			name:     "language is case-sensitive",
			trigger:  Trigger{Language: "Go"},
			wantErr:  ErrInvalidTrigger,
			contains: `Language "Go"`,
		},
		{
			name:     "empty Files pattern",
			trigger:  Trigger{Files: []string{"**/*.go", ""}},
			wantErr:  ErrInvalidTrigger,
			contains: "Files contains an empty pattern",
		},
		{
			name:     "empty Requires pattern",
			trigger:  Trigger{Requires: []string{""}},
			wantErr:  ErrInvalidTrigger,
			contains: "Requires contains an empty pattern",
		},
		{
			name:     "empty NotRequires pattern",
			trigger:  Trigger{NotRequires: []string{""}},
			wantErr:  ErrInvalidTrigger,
			contains: "NotRequires contains an empty pattern",
		},
		{
			name: "valid full trigger",
			trigger: Trigger{
				Files:       []string{"**/*.nix"},
				Language:    "nix",
				Requires:    []string{"**/flake.nix"},
				NotRequires: []string{"**/treefmt.toml"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.trigger.Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}

				return
			}

			if !errors.Is(err, tt.wantErr) || !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("Validate() = %v, want %v containing %q", err, tt.wantErr, tt.contains)
			}
		})
	}
}

func TestValidLanguagesListsDocumentedSet(t *testing.T) {
	t.Parallel()

	want := []string{"go", "js", "ts", "python", "nix", "rust"}

	got := ValidLanguages()
	if !slices.Equal(got, want) {
		t.Fatalf("ValidLanguages() = %v, want %v", got, want)
	}

	got[0] = "mutated"

	if ValidLanguages()[0] != "go" {
		t.Error("ValidLanguages() must return a copy immune to caller mutation")
	}
}

func TestRegisterRejectsInvalidTriggerLanguage(t *testing.T) {
	ResetForTest()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal(`Register with Language "golang" must panic`)
		}

		msg, ok := r.(string)
		if !ok {
			t.Fatalf("panic = %v, want string message", r)
		}

		for _, want := range []string{
			"Spec typo-tool",
			`Language "golang"`,
			"go, js, ts, python, nix, rust",
		} {
			if !strings.Contains(msg, want) {
				t.Errorf("panic message %q must contain %q", msg, want)
			}
		}
	}()

	Register(Spec{
		Name:        "typo-tool",
		Description: "must fail registration",
		Trigger:     Trigger{Language: "golang"},
		Detect: finding.NamedDetectorFunc(
			"typo-tool",
			func(context.Context) ([]finding.Finding, error) { return nil, nil },
		),
	})
}

func TestRegisterRejectsEmptyTriggerPattern(t *testing.T) {
	ResetForTest()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("Register with an empty Files pattern must panic")
		}

		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "Files contains an empty pattern") {
			t.Fatalf("panic = %v, want message naming Files empty pattern", r)
		}
	}()

	Register(Spec{
		Name:        "empty-pattern-tool",
		Description: "must fail registration",
		Trigger:     Trigger{Files: []string{""}},
		Detect: finding.NamedDetectorFunc(
			"empty-pattern-tool",
			func(context.Context) ([]finding.Finding, error) { return nil, nil },
		),
	})
}

// TestRepairResultCarriesFactsAndModifiedFiles pins the provenance surface
// (issue #42): ModifiedFiles records what the repair wrote, Facts carry
// tool-defined key/value provenance, and both stay non-gating — fix counts
// remain measured by re-detection, never self-reported.
func TestRepairResultCarriesFactsAndModifiedFiles(t *testing.T) {
	t.Parallel()

	result := RepairResult{
		Description:   "migrated 2 go.mod files",
		ModifiedFiles: []string{"go.mod", "tools/go.mod"},
		Facts: []Fact{
			{Key: "baseline-failed", Value: "true"},
			{Key: "failure-cache-suppressed", Value: "true"},
		},
	}

	if result.Description != "migrated 2 go.mod files" {
		t.Errorf("Description = %q", result.Description)
	}

	if want := []string{"go.mod", "tools/go.mod"}; !slices.Equal(result.ModifiedFiles, want) {
		t.Errorf("ModifiedFiles = %v, want %v", result.ModifiedFiles, want)
	}

	if len(result.Facts) != 2 {
		t.Fatalf("Facts length = %d, want 2", len(result.Facts))
	}

	if result.Facts[0].Key != "baseline-failed" || result.Facts[0].Value != "true" {
		t.Errorf("Facts[0] = %+v, want baseline-failed=true", result.Facts[0])
	}

	if result.Facts[1].Key != "failure-cache-suppressed" || result.Facts[1].Value != "true" {
		t.Errorf("Facts[1] = %+v, want failure-cache-suppressed=true", result.Facts[1])
	}

	var zero RepairResult
	if zero.ModifiedFiles != nil || zero.Facts != nil {
		t.Error("zero-value RepairResult must have nil ModifiedFiles and Facts (additive fields only)")
	}
}
