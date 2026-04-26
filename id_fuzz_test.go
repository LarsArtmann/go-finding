package finding

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func FuzzGenerateID(f *testing.F) {
	f.Fuzz(func(t *testing.T, tool, rule, file string, line, col int) {
		id := GenerateID(tool, rule, Position{File: file, Line: line, Column: col})
		require.NotEmpty(t, id)

		parts := strings.Split(id, ":")
		require.GreaterOrEqual(t, len(parts), 2, "ID too short: %q", id)

		if tool != "" {
			assert.Equal(t, tool, parts[0])
		}

		if rule != "" {
			assert.Equal(t, rule, parts[1])
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
		require.True(t, p.OK(), "ParseID failed for generated ID %q", id)

		if tool != "" {
			assert.Equal(t, tool, p.Tool)
		}

		if rule != "" {
			assert.Equal(t, rule, p.Rule)
		}

		if line > 0 {
			assert.Equal(t, line, p.Line)
		}

		if col > 0 {
			assert.Equal(t, col, p.Column)
		}
	})
}

func FuzzIsHashID(f *testing.F) {
	f.Fuzz(func(t *testing.T, id string) {
		result := IsHashID(id)

		parts := strings.Split(id, ":")
		if len(parts) == 3 && len(parts[2]) == hashLength {
			assert.True(t, result, "IsHashID should return true for %q", id)
		}
	})
}
