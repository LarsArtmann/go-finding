package finding

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestIsAutoFixable_ValidateAgreement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		fixStrategy FixStrategy
		beforeCode  string
		afterCode   string
		autoFixable bool
		validFix    bool
	}{
		{"none no code", FixStrategyNone, "", "", false, true},
		{"suggest no code", FixStrategySuggest, "", "", false, true},
		{"direct no code", FixStrategyDirect, "", "", false, false},
		{"direct after only", FixStrategyDirect, "", "new", true, true},
		{"direct before only", FixStrategyDirect, "old", "", true, true},
		{"direct both", FixStrategyDirect, "old", "new", true, true},
		{"suggest both", FixStrategySuggest, "old", "new", false, true},
		{"ai no code", FixStrategyAI, "", "", false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			f := NewFinding("r", "t", "m", SeverityError, Pos("a.go", 1, 1), 0.5)
			f.FixStrategy = tc.fixStrategy
			f.BeforeCode = tc.beforeCode
			f.AfterCode = tc.afterCode

			g.Expect(f.IsAutoFixable()).To(Equal(tc.autoFixable))
			g.Expect(f.Validate() == nil).To(Equal(tc.validFix))
		})
	}
}

func TestFinding_Validate(t *testing.T) {
	t.Parallel()

	t.Run("valid finding", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		f := NewFinding("r", "t", "m", SeverityError, Pos("a.go", 1, 1), 0.5)
		g.Expect(f.Validate()).NotTo(HaveOccurred())
	})

	t.Run("missing required fields", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		f := Finding{}
		err := f.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(IsCategory(err, ErrCategoryValidation)).To(BeTrue())
	})

	t.Run("invalid severity", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		f := NewFinding("r", "t", "m", Severity("bogus"), Pos("a.go", 1, 1), 0.5)
		err := f.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(ContainSubstring("Severity")))
	})

	t.Run("invalid fix strategy", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		f := NewFinding("r", "t", "m", SeverityError, Pos("a.go", 1, 1), 0.5)
		f.FixStrategy = FixStrategy("bogus")
		err := f.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(ContainSubstring("FixStrategy")))
	})

	t.Run("confidence out of range", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		f := Finding{
			ID: "1", Rule: "r", ToolName: "t", Message: "m",
			Severity: SeverityError, Position: Pos("a.go", 1, 1),
			Confidence: 1.5,
		}
		err := f.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(ContainSubstring("Confidence")))
	})

	t.Run("negative confidence", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		f := Finding{
			ID: "1", Rule: "r", ToolName: "t", Message: "m",
			Severity: SeverityError, Position: Pos("a.go", 1, 1),
			Confidence: -0.5,
		}
		err := f.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(ContainSubstring("Confidence")))
	})

	t.Run("direct fix with afterCode only is valid", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		f := NewFinding("r", "t", "m", SeverityError, Pos("a.go", 1, 1), 0.5)
		f.FixStrategy = FixStrategyDirect
		f.AfterCode = "new"
		g.Expect(f.Validate()).NotTo(HaveOccurred())
	})

	t.Run("direct fix without beforeCode or afterCode is invalid", func(t *testing.T) {
		t.Parallel()
		g := NewWithT(t)
		f := NewFinding("r", "t", "m", SeverityError, Pos("a.go", 1, 1), 0.5)
		f.FixStrategy = FixStrategyDirect
		err := f.Validate()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err).To(MatchError(ContainSubstring("BeforeCode or AfterCode")))
	})
}

func TestFinding_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		f    Finding
		want bool
	}{
		{"valid", NewFinding("r", "t", "m", SeverityError, Pos("a.go", 1, 1), 0), true},
		{
			"empty id",
			Finding{
				Rule:     "r",
				ToolName: "t",
				Message:  "m",
				Severity: SeverityError,
				Position: Pos("a.go", 1, 1),
			},
			false,
		},
		{
			"empty rule",
			Finding{
				ID:       "1",
				ToolName: "t",
				Message:  "m",
				Severity: SeverityError,
				Position: Pos("a.go", 1, 1),
			},
			false,
		},
		{
			"empty tool",
			Finding{
				ID:       "1",
				Rule:     "r",
				Message:  "m",
				Severity: SeverityError,
				Position: Pos("a.go", 1, 1),
			},
			false,
		},
		{
			"empty message",
			Finding{
				ID:       "1",
				Rule:     "r",
				ToolName: "t",
				Severity: SeverityError,
				Position: Pos("a.go", 1, 1),
			},
			false,
		},
		{
			"invalid position",
			Finding{
				ID:       "1",
				Rule:     "r",
				ToolName: "t",
				Message:  "m",
				Severity: SeverityError,
				Position: Position{File: "a.go", Line: -1},
			},
			false,
		},
		{
			"invalid severity",
			Finding{
				ID:       "1",
				Rule:     "r",
				ToolName: "t",
				Message:  "m",
				Severity: "bogus",
				Position: Pos("a.go", 1, 1),
			},
			false,
		},
	}

	runIsValidTests(t, tests)
}
