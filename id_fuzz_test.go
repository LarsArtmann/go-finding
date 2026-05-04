package finding

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func FuzzGenerateID(f *testing.F) {
	f.Fuzz(func(t *testing.T, tool, rule, file string, line, col int) {
		g := NewWithT(t)
		id := GenerateID(tool, rule, Position{File: file, Line: line, Column: col})
		g.Expect(id).NotTo(BeEmpty())

		parts := strings.Split(id, ":")
		g.Expect(len(parts)).To(BeNumerically(">=", 2))

		if tool != "" {
			g.Expect(parts[0]).To(Equal(tool))
		}

		if rule != "" {
			g.Expect(parts[1]).To(Equal(rule))
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
		g := NewWithT(t)
		if line < 0 || col < 0 {
			t.Skip()
		}

		id := GenerateID(tool, rule, Position{File: file, Line: line, Column: col})

		p := ParseID(id)
		g.Expect(p.OK()).To(BeTrue())

		if tool != "" {
			g.Expect(p.Tool).To(Equal(tool))
		}

		if rule != "" {
			g.Expect(p.Rule).To(Equal(rule))
		}

		if line > 0 {
			g.Expect(p.Line).To(Equal(line))
		}

		if col > 0 {
			g.Expect(p.Column).To(Equal(col))
		}
	})
}

func FuzzIsHashID(f *testing.F) {
	f.Fuzz(func(t *testing.T, id string) {
		g := NewWithT(t)
		result := IsHashID(id)

		parts := strings.Split(id, ":")
		if len(parts) == 3 && len(parts[2]) == hashLength {
			g.Expect(result).To(BeTrue())
		}
	})
}
