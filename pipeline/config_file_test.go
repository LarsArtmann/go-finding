package pipeline

import (
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
		g := NewParallelGomega(t)

		cf := ConfigFile{}
		sev, ok := cf.SeverityFilter()
		g.Expect(ok).To(BeFalse())
		g.Expect(sev).To(Equal(finding.SeverityInfo))
	})

	t.Run("valid severity returns true", func(t *testing.T) {
		g := NewParallelGomega(t)

		cf := ConfigFile{Severity: "error"}
		sev, ok := cf.SeverityFilter()
		g.Expect(ok).To(BeTrue())
		g.Expect(sev).To(Equal(finding.SeverityError))
	})

	t.Run("invalid severity returns false", func(t *testing.T) {
		g := NewParallelGomega(t)

		cf := ConfigFile{Severity: "bogus"}
		_, ok := cf.SeverityFilter()
		g.Expect(ok).To(BeFalse())
	})
}

// TestConfigFile_ResolveProviders covers the ResolveProviders method.
func TestConfigFile_ResolveProviders(t *testing.T) {
	t.Parallel()

	t.Run("empty names returns nil", func(t *testing.T) {
		g := NewParallelGomega(t)

		cf := ConfigFile{}
		providers, err := cf.ResolveProviders(nil)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(providers).To(BeNil())
	})

	t.Run("known providers resolved", func(t *testing.T) {
		g := NewParallelGomega(t)

		offset := OffsetProvider{}
		cf := ConfigFile{ProviderNames: []string{offset.Name()}}
		providers, err := cf.ResolveProviders(map[string]FixProvider{
			offset.Name(): offset,
		})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(providers).To(HaveLen(1))
	})

	t.Run("unknown provider returns error", func(t *testing.T) {
		g := NewParallelGomega(t)

		cf := ConfigFile{ProviderNames: []string{"ghost"}}
		_, err := cf.ResolveProviders(map[string]FixProvider{})
		g.Expect(err).To(HaveOccurred())
		g.Expect(errors.Is(err, errUnknownProvider)).To(BeTrue())
	})
}

// TestConfigFile_ConfigFromReader_Error covers the reader decode error path.
func TestConfigFile_ConfigFromReader_Error(t *testing.T) {
	g := NewParallelGomega(t)

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
			g := NewParallelGomega(t)

			_, err := ConfigFromFile([]byte(tt.input))
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring(tt.errContains))
		})
	}
}
