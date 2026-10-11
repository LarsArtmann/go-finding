// Command multi-edit demonstrates the typed edit-list fix flow (issue #36):
//
//  1. A Finding carries Finding.Edits ([]finding.TextEdit) — the
//     authoritative machine representation of a fix that touches more than
//     one place (e.g. rename a declaration and every reference).
//  2. EditListProvider — first in the default provider chain — resolves the
//     list to byte-level edits: offsets directly, line/column via the shared
//     line index. Cross-file lists fail with ErrEditCrossFile; stale
//     offsets fail with ErrEditStale.
//  3. FixEngine applies all edits of all findings in one pass and counts
//     applied fixes per finding (a multi-edit fix counts once).
//  4. BeforeCode/AfterCode remain the human display summary and are
//     intentionally not validated against Edits (ADR #19).
//
// Run it from the repository root:
//
//	go run ./pipeline/examples/multi-edit
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/larsartmann/go-finding"
	"github.com/larsartmann/go-finding/pipeline"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	const demoContent = "package demo\n\nvar useLegacy = true\n\nfunc old() bool { return useLegacy }\n"
	content := []byte(demoContent)

	// All offsets are into the ORIGINAL content — the engine applies every
	// edit against one content snapshot in a single descending-offset pass.
	declStart := strings.Index(demoContent, "useLegacy")
	boolStart := strings.Index(demoContent, "true")
	bodyStart := strings.LastIndex(demoContent, "useLegacy")

	// A fix touching three places: rename the declaration, flip the value,
	// and rename the reference in the body — one finding, one fix, three
	// edits.
	rename := finding.Finding{
		ID:          "demo/multi-edit/1",
		Rule:        "legacy-flag",
		ToolName:    "demo-tool",
		Message:     "rename useLegacy to cacheEnabled and update every reference",
		Severity:    finding.SeverityWarning,
		Position:    finding.Position{File: "demo.go"},
		FixStrategy: finding.FixStrategyDirect,
		BeforeCode:  "useLegacy",
		AfterCode:   "cacheEnabled",
		Edits: []finding.TextEdit{
			replace("demo.go", declStart, len("useLegacy"), "cacheEnabled"),
			replace("demo.go", boolStart, len("true"), "false"),
			replace("demo.go", bodyStart, len("useLegacy"), "cacheEnabled"),
		},
	}

	engine := pipeline.NewFixEngine()
	result, applied, count := engine.Apply(content, []finding.Finding{rename})
	if count != 1 {
		return fmt.Errorf("expected 1 applied fix, got %d", count)
	}

	fmt.Printf("applied findings : %d\n", count)
	fmt.Printf("findings touched: %d\n", len(applied))
	fmt.Printf("result:\n%s\n", result)

	return nil
}

// replace builds a TextEdit replacing the span [offset, offset+len(oldText))
// with newText.
func replace(file string, offset, length int, newText string) finding.TextEdit {
	return finding.TextEdit{
		Start:   finding.Position{File: finding.FilePath(file), Offset: offset},
		End:     finding.Position{File: finding.FilePath(file), Offset: offset + length},
		NewText: newText,
	}
}
