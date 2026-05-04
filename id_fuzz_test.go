package finding

import (
	"strconv"
	"strings"
	"testing"
)

func FuzzGenerateID(f *testing.F) {
	f.Fuzz(func(t *testing.T, tool, rule, file string, line, col int) {
		id := GenerateID(tool, rule, Position{File: file, Line: line, Column: col})
		if id == "" {
			t.Fatal("GenerateID returned empty")
		}

		parts := strings.Split(id, ":")
		if len(parts) < 2 {
			t.Fatalf("ID has fewer than 2 parts: %q", id)
		}

		// Tool and rule may contain colons, so exact part matching is not reliable.
		// Just verify the ID is non-empty and parseable.
		if !strings.Contains(id, tool) && tool != "" {
			t.Fatalf("ID %q should contain tool %q", id, tool)
		}
	})
}

func FuzzParseID(f *testing.F) {
	f.Fuzz(func(t *testing.T, id string) {
		_ = ParseID(id) // must not panic
	})
}

func FuzzRoundTripID(f *testing.F) {
	f.Fuzz(func(t *testing.T, tool, rule, file string, line, col int) {
		if line < 0 || col < 0 {
			t.Skip()
		}

		// When line=0, GenerateID uses hash format and drops column/line info.
		if line == 0 {
			t.Skip()
		}

		// Skip inputs where tool, rule, or file contain colons — round-trip
		// is lossy for such inputs since ID format uses colon separators.
		if strings.Contains(tool, ":") || strings.Contains(rule, ":") || strings.Contains(file, ":") {
			t.Skip()
		}

		// Empty tool produces IDs that ParseID can't round-trip (OK() returns false).
		if tool == "" {
			t.Skip()
		}

		// Numeric-only file names are ambiguous with position data in the ID format.
		if _, err := strconv.Atoi(file); err == nil && file != "" {
			t.Skip()
		}

		id := GenerateID(tool, rule, Position{File: file, Line: line, Column: col})

		p := ParseID(id)
		if !p.OK() {
			t.Fatalf("ParseID(%q) failed", id)
		}

		if tool != "" && p.Tool != tool {
			t.Fatalf("Tool: got %q, want %q", p.Tool, tool)
		}

		if rule != "" && p.Rule != rule {
			t.Fatalf("Rule: got %q, want %q", p.Rule, rule)
		}

		if line > 0 && p.Line != line {
			t.Fatalf("Line: got %d, want %d", p.Line, line)
		}

		if col > 0 && p.Column != col {
			t.Fatalf("Column: got %d, want %d", p.Column, col)
		}
	})
}

func FuzzIsHashID(f *testing.F) {
	f.Fuzz(func(t *testing.T, id string) {
		_ = IsHashID(id) // must not panic
	})
}
