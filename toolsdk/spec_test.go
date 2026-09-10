package toolsdk

import (
	"context"
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
