package toolsdk

import (
	"slices"
	"testing"
)

// TestOnGoModule_RequiresUseGlobPatterns verifies that OnGoModule() uses glob
// patterns ("**/go.mod") rather than literal paths ("go.mod"). Literal paths
// only match at the project root, missing workspace sub-modules where go.mod
// lives in a subdirectory. This test catches the regression where OnGoModule()
// was changed from glob to literal patterns.
func TestOnGoModule_RequiresUseGlobPatterns(t *testing.T) {
	t.Parallel()

	trigger := OnGoModule()

	if len(trigger.Requires) != 2 {
		t.Fatalf("expected 2 Requires patterns, got %d: %v", len(trigger.Requires), trigger.Requires)
	}

	for _, req := range trigger.Requires {
		if !slices.Contains([]string{"**/go.mod", "**/go.work"}, req) {
			t.Errorf(
				"OnGoModule().Requires contains %q — must be a glob pattern (**/go.mod or **/go.work), "+
					"not a literal path. Literal paths miss workspace sub-modules.",
				req,
			)
		}
	}
}

// TestOnGoModule_FilesAndLanguage verifies the basic trigger shape.
func TestOnGoModule_FilesAndLanguage(t *testing.T) {
	t.Parallel()

	trigger := OnGoModule()

	if trigger.Language != "go" {
		t.Errorf("expected Language \"go\", got %q", trigger.Language)
	}

	if !slices.Contains(trigger.Files, "**/*.go") {
		t.Errorf("expected Files to contain \"**/*.go\", got %v", trigger.Files)
	}
}
