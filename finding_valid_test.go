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
			g := NewParallelGomega(t)

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

	tests := []struct {
		name        string
		mutate      func(*Finding)
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid finding",
			mutate:  func(*Finding) {},
			wantErr: false,
		},
		{
			name: "missing required fields",
			mutate: func(f *Finding) {
				*f = Finding{}
			},
			wantErr: true,
		},
		{
			name: "invalid severity",
			mutate: func(f *Finding) {
				f.Severity = Severity("bogus")
			},
			wantErr:     true,
			errContains: "Severity",
		},
		{
			name: "invalid fix strategy",
			mutate: func(f *Finding) {
				f.FixStrategy = FixStrategy("bogus")
			},
			wantErr:     true,
			errContains: "FixStrategy",
		},
		{
			name: "confidence out of range",
			mutate: func(f *Finding) {
				f.Confidence = 1.5
			},
			wantErr:     true,
			errContains: "Confidence",
		},
		{
			name: "negative confidence",
			mutate: func(f *Finding) {
				f.Confidence = -0.5
			},
			wantErr:     true,
			errContains: "Confidence",
		},
		{
			name: "direct fix with afterCode only is valid",
			mutate: func(f *Finding) {
				f.FixStrategy = FixStrategyDirect
				f.AfterCode = "new"
			},
			wantErr: false,
		},
		{
			name: "direct fix without beforeCode or afterCode is invalid",
			mutate: func(f *Finding) {
				f.FixStrategy = FixStrategyDirect
			},
			wantErr:     true,
			errContains: "BeforeCode or AfterCode",
		},
		{
			name:    "nil range is valid",
			mutate:  func(*Finding) {},
			wantErr: false,
		},
		{
			name: "valid range passes",
			mutate: func(f *Finding) {
				f.Range = &Range{Start: Pos("a.go", 1, 1), End: Pos("a.go", 3, 1)}
			},
			wantErr: false,
		},
		{
			name: "inverted range is invalid",
			mutate: func(f *Finding) {
				f.Range = &Range{Start: Pos("a.go", 5, 1), End: Pos("a.go", 1, 1)}
			},
			wantErr:     true,
			errContains: "Range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewParallelGomega(t)

			f := NewFinding("r", "t", "m", SeverityError, Pos("a.go", 1, 1), 0.5)
			tt.mutate(&f)

			err := f.Validate()
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(IsCategory(err, ErrCategoryValidation)).To(BeTrue())

				if tt.errContains != "" {
					g.Expect(err).To(MatchError(ContainSubstring(tt.errContains)))
				}
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
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
				Position: Position{},
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

func TestValidateAll(t *testing.T) {
	t.Parallel()

	validFinding := NewFinding("rule1", "tool", "msg", SeverityError, Pos("a.go", 1, 1), 0.5)
	invalidFinding := Finding{}

	tests := []struct {
		name     string
		findings []Finding
		wantNil  bool
		wantKeys []int
	}{
		{"nil slice", nil, true, nil},
		{"empty slice", []Finding{}, true, nil},
		{"all valid", []Finding{validFinding, validFinding}, true, nil},
		{"all invalid", []Finding{invalidFinding, invalidFinding}, false, []int{0, 1}},
		{"mixed valid and invalid", []Finding{validFinding, invalidFinding}, false, []int{1}},
		{"single invalid at index 0", []Finding{invalidFinding}, false, []int{0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewParallelGomega(t)

			result := ValidateAll(tt.findings)
			if tt.wantNil {
				g.Expect(result).To(BeNil())
			} else {
				g.Expect(result).NotTo(BeNil())
				g.Expect(result).To(HaveLen(len(tt.wantKeys)))

				for _, key := range tt.wantKeys {
					g.Expect(result).To(HaveKey(key))
					g.Expect(result[key]).To(HaveOccurred())
				}
			}
		})
	}
}
