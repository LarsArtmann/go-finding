package pipeline

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
	. "github.com/onsi/gomega"
)

// TestConfigFile_SeverityFilter covers the SeverityFilter method.
func TestConfigFile_SeverityFilter(t *testing.T) {
	t.Parallel()

	t.Run("empty returns false", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		cf := ConfigFile{}
		sev, ok := cf.SeverityFilter()
		g.Expect(ok).To(BeFalse())
		g.Expect(sev).To(Equal(finding.SeverityInfo))
	})

	t.Run("valid severity returns true", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		cf := ConfigFile{Severity: "error"}
		sev, ok := cf.SeverityFilter()
		g.Expect(ok).To(BeTrue())
		g.Expect(sev).To(Equal(finding.SeverityError))
	})

	t.Run("invalid severity returns false", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		cf := ConfigFile{Severity: "bogus"}
		_, ok := cf.SeverityFilter()
		g.Expect(ok).To(BeFalse())
	})
}

// TestConfigFile_ResolveProviders covers the ResolveProviders method.
func TestConfigFile_ResolveProviders(t *testing.T) {
	t.Parallel()

	t.Run("empty names returns nil", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		cf := ConfigFile{}
		providers, err := cf.ResolveProviders(nil)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(providers).To(BeNil())
	})

	t.Run("known providers resolved", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		offset := OffsetProvider{}
		cf := ConfigFile{ProviderNames: []string{offset.Name()}}
		providers, err := cf.ResolveProviders(map[string]FixProvider{
			offset.Name(): offset,
		})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(providers).To(HaveLen(1))
	})

	t.Run("unknown provider returns error", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)

		cf := ConfigFile{ProviderNames: []string{"ghost"}}
		_, err := cf.ResolveProviders(map[string]FixProvider{})
		g.Expect(err).To(HaveOccurred())
		g.Expect(errors.Is(err, errUnknownProvider)).To(BeTrue())
	})
}

// TestConfigFile_ConfigFromReader_Error covers the reader decode error path.
func TestConfigFile_ConfigFromReader_Error(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	_, err := ConfigFromReader(strings.NewReader("{invalid json"))
	g.Expect(err).To(HaveOccurred())
}

// TestConfigFile_ConfigFromFile_BadDurations covers every variant of bad
// duration strings accepted by ConfigFromFile: the top-level `timeout` field
// and the per-detector `detectorTimeouts` map. Each case is checked against
// the substring that the error message is expected to carry.
func TestConfigFile_ConfigFromFile_BadDurations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		errContains string
	}{
		{
			name:        "BadTimeout",
			input:       `{"timeout": "not-a-duration"}`,
			errContains: "parse timeout",
		},
		{
			name:        "BadDetectorTimeout",
			input:       `{"detectorTimeouts": {"govet": "bad"}}`,
			errContains: "parse detector timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			g := NewWithT(t)

			_, err := ConfigFromFile([]byte(tt.input))
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring(tt.errContains))
		})
	}
}

// TestFixApplier_ApplyWithDetails covers the ApplyWithDetails wrapper.
func TestFixApplier_ApplyWithDetails(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	applier, err := NewFixApplier(t.TempDir())
	g.Expect(err).NotTo(HaveOccurred())

	count, fixes, err := applier.ApplyWithDetails(context.Background(), nil)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(count).To(Equal(0))
	g.Expect(fixes).To(BeEmpty())
}

// TestPickNearestOccurrence_Branches covers the remaining branches of
// pickNearestOccurrence: line-only fallback, column path, and the
// line-out-of-range early return.
func TestPickNearestOccurrence_Branches(t *testing.T) {
	t.Parallel()

	content := []byte("X here\nX there\nX everywhere")
	idx := buildLineOffsetIndex(content)
	occurrences := findAllOccurrences(content, []byte("X")) // offsets 0, 7, 15

	t.Run("line only fallback", func(t *testing.T) {
		t.Parallel()

		// Target line 2; closest occurrence is offset 7 (line 2).
		f := finding.Finding{Position: finding.Position{Line: 2}}

		best := pickNearestOccurrence(idx, occurrences, f)
		if best != 7 {
			t.Fatalf("line-only: got %d, want 7", best)
		}
	})

	t.Run("line out of range returns first", func(t *testing.T) {
		t.Parallel()

		f := finding.Finding{Position: finding.Position{Line: 99}}

		best := pickNearestOccurrence(idx, occurrences, f)
		if best != occurrences[0] {
			t.Fatalf("out-of-range: got %d, want %d", best, occurrences[0])
		}
	})

	t.Run("column nearest", func(t *testing.T) {
		t.Parallel()

		// Two X's could be close; column helps disambiguate.
		// On line 1, column 1 → offset 0 is closest.
		f := finding.Finding{Position: finding.Position{Line: 1, Column: 1}}

		best := pickNearestOccurrence(idx, occurrences, f)
		if best != 0 {
			t.Fatalf("col1: got %d, want 0", best)
		}
	})
}
