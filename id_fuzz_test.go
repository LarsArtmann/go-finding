package finding

import (
	"strings"
	"testing"
)

func FuzzGenerateID(f *testing.F) {
	f.Fuzz(func(t *testing.T, tool, rule, file string, line, col int) {
		id := GenerateID(tool, rule, Position{File: file, Line: line, Column: col})
		if id == "" {
			t.Fatal("GenerateID returned empty string")
		}

		parts := strings.Split(id, ":")
		if len(parts) < 2 {
			t.Fatalf("ID too short: %q", id)
		}

		if tool != "" && parts[0] != tool {
			t.Fatalf("tool mismatch: got %q, want %q", parts[0], tool)
		}

		if rule != "" && parts[1] != rule {
			t.Fatalf("rule mismatch: got %q, want %q", parts[1], rule)
		}
	})
}

func FuzzParseID(f *testing.F) {
	f.Fuzz(func(t *testing.T, id string) {
		p := ParseID(id)
		if !p.OK() {
			if len(strings.Split(id, ":")) >= 3 {
				t.Fatalf("ParseID should succeed for multi-part ID %q", id)
			}

			return
		}

		if p.Tool == "" && len(strings.Split(id, ":")) >= 3 {
			t.Fatalf("ParseID returned empty tool for %q", id)
		}
	})
}

func FuzzRoundTripID(f *testing.F) {
	f.Fuzz(func(t *testing.T, tool, rule, file string, line, col int) {
		if line < 0 || col < 0 {
			t.Skip()
		}

		id := GenerateID(tool, rule, Position{File: file, Line: line, Column: col})

		p := ParseID(id)
		if !p.OK() {
			t.Fatalf("ParseID failed for generated ID %q", id)
		}

		parsedTool, parsedRule, parsedLine, parsedCol := p.Tool, p.Rule, p.Line, p.Column

		if tool != "" && parsedTool != tool {
			t.Fatalf("tool round-trip mismatch: got %q, want %q", parsedTool, tool)
		}

		if rule != "" && parsedRule != rule {
			t.Fatalf("rule round-trip mismatch: got %q, want %q", parsedRule, rule)
		}

		if line > 0 && parsedLine != line {
			t.Fatalf("line round-trip mismatch: got %d, want %d", parsedLine, line)
		}

		if col > 0 && parsedCol != col {
			t.Fatalf("col round-trip mismatch: got %d, want %d", parsedCol, col)
		}
	})
}

func FuzzIsHashID(f *testing.F) {
	f.Fuzz(func(t *testing.T, id string) {
		result := IsHashID(id)

		parts := strings.Split(id, ":")
		if len(parts) == 3 && len(parts[2]) == hashLength {
			if !result {
				t.Fatalf("IsHashID should return true for %q", id)
			}
		}
	})
}
