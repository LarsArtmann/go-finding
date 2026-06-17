package finding

import (
	"errors"
	"fmt"
)

// Validate checks all fields of the Finding for correctness and returns detailed
// per-field errors. Use IsValid for a simple boolean check.
func (f Finding) Validate() error {
	var errs []error

	if f.ID == "" {
		errs = append(errs, NewValidationError("finding.ID is required", nil))
	}

	if f.Rule == "" {
		errs = append(errs, NewValidationError("finding.Rule is required", nil))
	}

	if f.ToolName == "" {
		errs = append(errs, NewValidationError("finding.ToolName is required", nil))
	}

	if f.Message == "" {
		errs = append(errs, NewValidationError("finding.Message is required", nil))
	}

	if !f.Severity.IsValid() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf("finding.Severity %q is invalid", f.Severity), nil,
		))
	}

	if !f.Position.IsValid() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf("finding.Position %+v is invalid", f.Position), nil,
		))
	}

	if f.FixStrategy == "" {
		f.FixStrategy = FixStrategyNone
	} else if !f.FixStrategy.IsValid() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf("finding.FixStrategy %q is invalid", f.FixStrategy), nil,
		))
	}

	if !f.Confidence.IsValid() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf("finding.Confidence %s must be in [0.0, 1.0]", f.Confidence), nil,
		))
	}

	if f.BeforeCode == "" && f.AfterCode == "" && f.FixStrategy == FixStrategyDirect {
		errs = append(errs, NewValidationError(
			"finding.FixStrategyDirect requires BeforeCode or AfterCode", nil,
		))
	}

	for i, tag := range f.Tags {
		if !tag.IsValid() {
			errs = append(errs, NewValidationError(
				fmt.Sprintf("finding.Tags[%d] %q is invalid: tags must be non-empty", i, tag), nil,
			))
		}
	}

	// Category/Tags consistency: if both Category and Tags are set, the Category
	// value must appear in Tags (or Tags must not contain a conflicting standard
	// category). This prevents the split-brain where Category says "security" but
	// Tags only contains "performance".
	if f.Category != "" && len(f.Tags) > 0 && f.Category.IsStandard() {
		categoryTag := Tag(f.Category)
		hasCategoryTag := false
		hasOtherStandardCategory := false

		for _, tag := range f.Tags {
			if tag == categoryTag {
				hasCategoryTag = true
			} else if tag.IsStandard() && Category(tag).IsStandard() &&
				Category(tag) != f.Category {
				hasOtherStandardCategory = true
			}
		}

		if !hasCategoryTag && hasOtherStandardCategory {
			errs = append(errs, NewValidationError(
				fmt.Sprintf(
					"finding.Category %q conflicts with Tags: when Category is a standard category "+
						"and Tags contain a different standard category, Category must also appear in Tags",
					f.Category,
				),
				nil,
			))
		}
	}

	for i, ref := range f.Related {
		if !ref.IsValid() {
			errs = append(errs, NewValidationError(
				fmt.Sprintf("finding.Related[%d] is invalid: missing required fields", i), nil,
			))
		}

		if ref.Range != nil && ref.Range.IsInverted() {
			errs = append(errs, NewValidationError(
				fmt.Sprintf(
					"finding.Related[%d].Range is invalid: End (%+v) before Start (%+v)",
					i, ref.Range.End, ref.Range.Start,
				),
				nil,
			))
		}
	}

	if f.Suppression != nil && !f.Suppression.IsValid() {
		errs = append(errs, NewValidationError(
			"finding.Suppression is invalid: missing Kind or Rule", nil,
		))
	}

	if f.Range != nil && f.Range.IsInverted() {
		errs = append(errs, NewValidationError(
			fmt.Sprintf(
				"finding.Range is invalid: End (%+v) before Start (%+v)",
				f.Range.End,
				f.Range.Start,
			),
			nil,
		))
	}

	return errors.Join(errs...)
}

// IsValid returns true if the finding has required fields set.
func (f Finding) IsValid() bool {
	return f.ID != "" && f.Rule != "" && f.ToolName != "" &&
		f.Message != "" && f.Position.IsValid() && f.Severity.IsValid()
}

// Key returns a stable identifier for the finding.
//
// Canonical identity: Two findings are identical iff their GenerateID outputs
// are equal (see ADR #12). Key() is a fallback for findings without an ID,
// building a composite key from ToolName, Position.File, Rule, and Message.
//
// Note: Key() includes Message in the fallback key, while GenerateID does not.
// This means two findings with different messages but the same position will
// have different Keys but the same GenerateID. Prefer ID (and thus GenerateID)
// as the canonical identity. Key() is primarily used by external dedup logic
// that predates GenerateID.
func (f Finding) Key() string {
	if f.ID != "" {
		return f.ID
	}

	return f.ToolName + keySeparator + f.Position.File + keySeparator + f.Rule + keySeparator + f.Message
}
